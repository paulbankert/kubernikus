package walle

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	api_v1 "k8s.io/client-go/pkg/api/v1"

	"github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
	"github.com/sapcc/kubernikus/pkg/controller/config"
)

const (
	BaseDelay              = 5 * time.Second
	MaxDelay               = 300 * time.Second
	KlusterRecheckInterval = 5 * time.Minute
)

type WalleController struct {
	Controller
	config.Factories
	config.Clients
	queue           workqueue.RateLimitingInterface
	logger          log.Logger
	nodeInformerMap sync.Map
	stopCh          <-chan struct{}
	informerWg      *sync.WaitGroup
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
	wr.informerWg = wg

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

	if err := wr.cleanUpInformers(); err != nil {
		return true, err
	}

	if err := wr.watchNodesForKluster(kluster); err != nil {
		return true, err
	}

	return false, nil
}

func (wr *WalleController) cleanUpInformers() error {
	wr.nodeInformerMap.Range(
		func(key, value interface{}) bool {
			if _, exists, _ := wr.Factories.Kubernikus.Kubernikus().V1().Klusters().Informer().GetIndexer().GetByKey(key.(string)); !exists {
				wr.nodeInformerMap.Delete(key)
			}
			return true
		},
	)
	return nil
}

func (wr *WalleController) createNodeInformerForKluster(kluster *v1.Kluster) error {
	key, err := cache.MetaNamespaceKeyFunc(kluster)
	if err != nil {
		return err
	}

	_, exists := wr.nodeInformerMap.Load(key)
	if !exists {
		nodeInformer, err := NewNodeInformerForKluster(wr.Clients.Satellites, kluster, make(chan struct{}))
		if err != nil {
			return err
		}

		nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				wr.logger.Log("added node %s", obj.(*api_v1.Node).GetName())
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				wr.logger.Log("updated node %s", oldObj.(*api_v1.Node).GetName())
			},
			DeleteFunc: func(obj interface{}) {
				wr.logger.Log("deleted node %s", obj.(*api_v1.Node).GetName())
			},
		})

		key, err := nodeInformer.Key()
		if err != nil {
			return err
		}

		wr.nodeInformerMap.Store(
			key,
			nodeInformer,
		)
	}
	return nil
}

func (wr *WalleController) watchNodesForKluster(kluster *v1.Kluster) error {
	key, err := cache.MetaNamespaceKeyFunc(kluster)
	if err != nil {
		return err
	}

	i, ok := wr.nodeInformerMap.Load(key)
	if !ok {
		return fmt.Errorf("no node informer created for kluster %s/%s", kluster.GetNamespace(), kluster.GetName())
	}

	informer := i.(*NodeInformer)
	cache.WaitForCacheSync(
		wr.stopCh,
		informer.HasSynced,
	)

	go func(informer *NodeInformer) {
		wr.informerWg.Add(1)
		defer wr.informerWg.Done()
		for {
			select {
			case <-wr.stopCh:
				informer.Close()
				return
			default:
				informer.Run()
			}
		}
	}(informer)

	return nil
}

func (wr *WalleController) klusterAddFunc(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		wr.queue.Add(key)
	}
}

func (wr *WalleController) klusterUpdateFunc(cur, old interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(cur)
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
		obj.(*NodeInformer).Close()
		wr.nodeInformerMap.Delete(key)
	}
}

func (wr *WalleController) GetNodeInformerForKluster(kluster *v1.Kluster) (NodeInformer, error) {
	key, err := cache.MetaNamespaceKeyFunc(kluster)
	if err != nil {
		return NodeInformer{}, err
	}
	i, isFound := wr.nodeInformerMap.Load(key)
	if !isFound {
		return NodeInformer{}, fmt.Errorf("no node informers found for kluster %s/%s", kluster.GetNamespace(), kluster.GetName())
	}
	return i.(NodeInformer), nil
}
