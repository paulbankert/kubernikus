package walle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	api_v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
	"github.com/sapcc/kubernikus/pkg/controller/config"
)

const (
	BaseDelay              = 5 * time.Second
	MaxDelay               = 300 * time.Second
	KlusterRecheckInterval = 5 * time.Minute
	NodeResyncPeriod       = 5 * time.Minute
)

type WalleController struct {
	//Controller
	config.Factories
	config.Clients
	queue           workqueue.RateLimitingInterface
	logger          log.Logger
	nodeInformerMap sync.Map
	stopCh          <-chan struct{}
}

func NewController(factories config.Factories, clients config.Clients, logger log.Logger) *WalleController {
	logger = log.With(logger,
		"controller", "walle",
	)

	controller := &WalleController{
		Factories:       factories,
		Clients:         clients,
		queue:           workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(BaseDelay, MaxDelay)),
		logger:          logger,
		nodeInformerMap: sync.Map{},
	}

	controller.Factories.Kubernikus.Kubernikus().V1().Klusters().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.klusterAddFunc,
		UpdateFunc: controller.klusterUpdateFunc,
		DeleteFunc: controller.klusterDeleteFunc,
	})

	return controller
}

func (wr *WalleController) Run(threadiness int, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	wr.logger.Log(
		"msg", "starting run loop",
		"threadiness", threadiness,
		"v", 2,
	)
	wr.stopCh = stopCh

	defer wr.queue.ShutDown()
	defer wg.Done()
	wg.Add(1)

	for i := 0; i < threadiness; i++ {
		go wait.Until(wr.runWorker, time.Second, stopCh)
	}

	ticker := time.NewTicker(KlusterRecheckInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				wr.requeueAllKlusters()
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
	<-stopCh
}

func (wr *WalleController) requeueAllKlusters() (err error) {
	defer func() {
		wr.logger.Log(
			"msg", "requeued all",
			"v", 1,
			"err", err,
		)
	}()

	klusters, err := wr.Factories.Kubernikus.Kubernikus().V1().Klusters().Lister().List(labels.Everything())
	if err != nil {
		return err
	}

	for _, kluster := range klusters {
		wr.requeueKluster(kluster)
	}

	return nil
}

func (wr *WalleController) requeueKluster(kluster *v1.Kluster) {
	wr.logger.Log(
		"msg", "queuing",
		"kluster", kluster.Spec.Name,
		"project", kluster.Account(),
		"v", 2,
	)

	key, err := cache.MetaNamespaceKeyFunc(kluster)
	if err == nil {
		wr.queue.Add(key)
	}
}

func (wr *WalleController) runWorker() {
	for wr.processNextWorkItem() {
	}
}

func (wr *WalleController) processNextWorkItem() bool {
	key, quit := wr.queue.Get()
	if quit {
		return false
	}
	defer wr.queue.Done(key)

	var err error
	var kluster *v1.Kluster
	var requeue bool

	obj, exists, _ := wr.Factories.Kubernikus.Kubernikus().V1().Klusters().Informer().GetIndexer().GetByKey(key.(string))
	if !exists {
		err = errors.NewNotFound(schema.GroupResource{Resource: "kluster"}, "")
	} else {
		kluster = obj.(*v1.Kluster)
	}

	if err == nil {
		// Invoke the method containing the business logic
		requeue, err = wr.Reconcile(kluster)
	}

	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.

		if requeue == false {
			wr.queue.Forget(key)
		} else {
			// Requeue requested
			wr.queue.AddAfter(key, BaseDelay)
		}

		return true
	}

	if wr.queue.NumRequeues(key) < 5 {
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		wr.queue.AddRateLimited(key)
		return true
	}

	// Retries exceeded. Forgetting for this reconciliation loop
	wr.queue.Forget(key)
	return true
}

func (wr *WalleController) Reconcile(kluster *v1.Kluster) (bool, error) {

	if err := wr.createNodeInformerForKluster(kluster); err != nil {
		return true, err
	}

	if err := wr.watchNodesForKluster(kluster); err != nil {
		return true, err
	}

	return false, nil
}

func (wr *WalleController) createNodeInformerForKluster(kluster *v1.Kluster) error {
	_, isFound := wr.nodeInformerMap.Load(klusterKey(kluster))
	if !isFound {
		client, err := wr.Clients.Satellites.ClientFor(kluster)
		if err != nil {
			return err
		}
		nodeInformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return client.CoreV1().Nodes().List(meta_v1.ListOptions{})
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return client.CoreV1().Nodes().Watch(meta_v1.ListOptions{})
				},
			},
			&api_v1.Node{},
			NodeResyncPeriod,
			cache.Indexers{},
		)

		nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    wr.nodeAddFunc,
			UpdateFunc: wr.nodeUpdateFunc,
			DeleteFunc: wr.nodeDeleteFunc,
		})

		wr.nodeInformerMap.Store(
			klusterKey(kluster),
			nodeInformer,
		)
	}
	return nil
}

func (wr *WalleController) watchNodesForKluster(kluster *v1.Kluster) error {

	i, ok := wr.nodeInformerMap.Load(klusterKey(kluster))
	if !ok {
		return fmt.Errorf("no node informer created for kluster %s/%s", kluster.GetNamespace(), kluster.GetName())
	}

	informer := i.(cache.SharedIndexInformer)
	cache.WaitForCacheSync(
		wr.stopCh,
		informer.HasSynced,
	)

	go func(ctx context.Context, informer cache.SharedIndexInformer) {
		informer.Run(wr.stopCh)
	}(context.Background(), informer)

	return nil
}

func (wr *WalleController) klusterAddFunc(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		wr.queue.Add(key)
	}
}

func (wr *WalleController) klusterUpdateFunc(cur, old interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err == nil {
		wr.queue.Add(key)
	}
}

func (wr *WalleController) klusterDeleteFunc(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		wr.queue.Add(key)
	}
	obj, exists, _ := wr.Factories.Kubernikus.Kubernikus().V1().Klusters().Informer().GetIndexer().GetByKey(key)
	if exists {
		wr.nodeInformerMap.Delete(klusterKey(obj.(*v1.Kluster)))
	}
}

func (wr *WalleController) GetNodeInformerMap() *sync.Map {
	return &wr.nodeInformerMap
}

func (wr *WalleController) GetNodeInformerForKluster(kluster *v1.Kluster) (cache.SharedIndexInformer, error) {
	i, isFound := wr.nodeInformerMap.Load(klusterKey(kluster))
	if !isFound {
		return nil, fmt.Errorf("no node informers found for kluster %s/%s", kluster.GetNamespace(), kluster.GetName())
	}
	return i.(cache.SharedIndexInformer), nil
}

func (wr *WalleController) nodeAddFunc(obj interface{}) {
	//TODO
}

func (wr *WalleController) nodeUpdateFunc(cur, old interface{}) {
	//TODO
}

func (wr *WalleController) nodeDeleteFunc(obj interface{}) {
	//TODO
}

func klusterKey(kluster *v1.Kluster) string {
	return fmt.Sprintf("%s/%s", kluster.GetNamespace(), kluster.GetName())
}
