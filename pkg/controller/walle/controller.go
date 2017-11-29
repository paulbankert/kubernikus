package walle

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
	"github.com/sapcc/kubernikus/pkg/controller/base"
	"github.com/sapcc/kubernikus/pkg/controller/config"

	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	api_v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/record"
)

const (
	NodeWatchTimeout = 5 * time.Minute
)

type WalleReconciler struct {
	config.Clients

	Recorder record.EventRecorder
	Logger   log.Logger

	nodeWatcherMap sync.Map
}

func NewController(factories config.Factories, clients config.Clients, recorder record.EventRecorder, logger log.Logger) base.Controller {
	logger = log.With(logger,
		"controller", "walle",
	)

	var reconciler base.Reconciler
	reconciler = &WalleReconciler{clients, recorder, logger, sync.Map{}}
	return base.NewController(factories, clients, reconciler, logger)
}

func (wr *WalleReconciler) Reconcile(kluster *v1.Kluster) (requeueRequested bool, err error) {

	requeueRequested, err = wr.createNodeWatcherForKluster(kluster)
	if err != nil {
		return true, err
	}

	go func(kluster *v1.Kluster) {
		requeueRequested, err = wr.watchNodesForKluster(kluster)
		if err != nil {
			wr.Logger.Log(err)
		}
	}(kluster)
	return requeueRequested, nil
}

// connect to kluster, list nodes, create watcher and store in sync.Map
func (wr *WalleReconciler) createNodeWatcherForKluster(kluster *v1.Kluster) (requeueRequested bool, err error) {

	client, err := wr.Clients.Satellites.ClientFor(kluster)
	if err != nil {
		return true, err
	}

	nodeList, err := client.CoreV1().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		return true, err
	}

	// store watchers as in map with map[string]watch.Inteface{ nodeName : watcher }
	var wm map[string]watch.Interface
	w, isFound := wr.nodeWatcherMap.Load(kluster.GetName())
	if isFound {
		wm = w.(map[string]watch.Interface)
	} else {
		wm = make(map[string]watch.Interface, len(nodeList.Items))
	}

	for _, node := range nodeList.Items {
		nodeName := node.GetName()
		// if watcher for node doesn't exist create
		if _, ok := wm[nodeName]; !ok {
			nodeWatcher, err := client.CoreV1().Nodes().Watch(
				meta_v1.SingleObject(
					meta_v1.ObjectMeta{
						Name: node.Name,
					},
				),
			)
			if err != nil {
				return true, err
			}
			wm[nodeName] = nodeWatcher
		}
	}
	wr.nodeWatcherMap.Store(kluster.GetName(), wm)
	return
}

//TODO: if kluster is deleted: cleanup wr.nodeWatcherMap
func (wr *WalleReconciler) klusterDeleted(klusterName string) {
	// cleanup map
	wr.nodeWatcherMap.Delete(klusterName)
}

func (wr *WalleReconciler) nodeDelete(klusterName, nodeName string) {
	wm, ok := wr.nodeWatcherMap.Load(klusterName)
	if ok {
		watcherMap := wm.(map[string]watch.Interface)
		delete(watcherMap, nodeName)
		wr.nodeWatcherMap.Store(klusterName, watcherMap)
	}
}

func (wr *WalleReconciler) updateNodeStatus(kluster *v1.Kluster, nodeName string, nodeStatusPhase api_v1.NodePhase) {
	wr.Logger.Log("updating status of node %s: %v", nodeName, nodeStatusPhase)
	//TODO: update in new kluster status field?
}

func (wr *WalleReconciler) watchNodesForKluster(kluster *v1.Kluster) (requeueRequested bool, err error) {
	klusterName := kluster.GetName()
	nw, ok := wr.nodeWatcherMap.Load(klusterName)
	if !ok {
		return true, fmt.Errorf("no node watchers were created for kluster %s", klusterName)
	}

	nodeWatcher := nw.(map[string]watch.Interface)
	for nodeName, watcher := range nodeWatcher {
		wr.Logger.Log("watching node %s", nodeName)

		// only stop watching if node is terminated
		event, err := watch.Until(NodeWatchTimeout, watcher, wr.isNodeTerminated)
		if err != nil {
			if errors.IsNotFound(err) {
				wr.nodeDelete(klusterName, nodeName)
			}
			return true, err
		}
		switch node := event.Object.(type) {
		case *api_v1.Node:
			wr.updateNodeStatus(kluster, node.GetName(), node.Status.Phase)
		}
		return false, nil
	}
	return
}

func (wr *WalleReconciler) isNodeTerminated(event watch.Event) (bool, error) {
	switch event.Type {
	case watch.Deleted:
		return false, errors.NewNotFound(schema.GroupResource{Resource: "nodes"}, "")
	}
	switch n := event.Object.(type) {
	case *api_v1.Node:
		switch n.Status.Phase {
		case api_v1.NodeTerminated:
			return false, errors.NewGone(fmt.Sprintf("node %s was deleted", n.GetName()))
		}
	}
	return false, nil
}

func (wr *WalleReconciler) isKlusterNeedsUpdate(kCur, kOld *v1.Kluster) bool {
	if !reflect.DeepEqual(kOld, kCur) {
		return true
	}
	return false
}

func (wr *WalleReconciler) updateUpstreamKluster(kCur, kOld *v1.Kluster) error {
	if wr.isKlusterNeedsUpdate(kCur, kOld) {
		wr.Logger.Log("Updating kluster %s/%s", kOld.GetNamespace(), kOld.GetName())

		_, err := wr.Clients.Kubernikus.KubernikusV1().Klusters(kOld.GetNamespace()).UpdateStatus(kCur)
		if err != nil {
			return fmt.Errorf("Failed to update kluster %s/%s: %v", err)
		}
	}
	return nil
}
