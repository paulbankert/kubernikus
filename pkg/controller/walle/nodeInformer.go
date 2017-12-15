package walle

import (
	"fmt"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	api_v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/sapcc/kubernikus/pkg/client/kubernetes"

	"github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
)

const (
	NodeResyncPeriod = 5 * time.Minute
)

type NodeInformer struct {
	cache.SharedIndexInformer

	kluster *v1.Kluster
	stopCh  chan struct{}
}

func NewNodeInformer(informer cache.SharedIndexInformer, kluster *v1.Kluster, stopCh chan struct{}) *NodeInformer {
	return &NodeInformer{
		SharedIndexInformer: informer,
		kluster:             kluster,
		stopCh:              stopCh,
	}
}

func NewNodeInformerForKluster(clientFactory *kubernetes.SharedClientFactory, kluster *v1.Kluster, stopCh chan struct{}) (*NodeInformer, error) {
	client, err := clientFactory.ClientFor(kluster)
	if err != nil {
		return nil, err
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

	return &NodeInformer{
		SharedIndexInformer: nodeInformer,
		kluster:             kluster,
		stopCh:              stopCh,
	}, nil
}

func (ni *NodeInformer) Run() {
	ni.SharedIndexInformer.Run(ni.stopCh)
}

func (ni *NodeInformer) Close() {
	close(ni.stopCh)
}

func (ni *NodeInformer) GetKluster() *v1.Kluster {
	return ni.kluster
}

func (ni *NodeInformer) Key() (string, error) {
	return cache.MetaNamespaceKeyFunc(ni.kluster)
}

func (ni *NodeInformer) GetNodeByKey(key string) (*api_v1.Node, error) {
	obj, exists, err := ni.SharedIndexInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("node %s not found", key)
	}
	return obj.(*api_v1.Node), nil
}

func (ni *NodeInformer) AddEventHandlerAddFunc(addFunc func(obj interface{})) {
	ni.SharedIndexInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: addFunc,
		},
	)
}

func (ni *NodeInformer) AddEventHandlerUpdateFunc(updateFunc func(cur, old interface{})) {
	ni.SharedIndexInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: updateFunc,
		},
	)
}

func (ni *NodeInformer) AddEventHandlerDeleteFunc(deleteFunc func(obj interface{})) {
	ni.SharedIndexInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: deleteFunc,
		},
	)
}
