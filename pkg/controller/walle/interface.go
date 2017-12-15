package walle

import (
	"sync"

	api_v1 "k8s.io/client-go/pkg/api/v1"

	"github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
	"github.com/sapcc/kubernikus/pkg/client/kubernetes"
)

type Controller interface {
	Run(threadiness int, stopCh <-chan struct{}, wg *sync.WaitGroup)
	TearDown()
}

type Informer interface {
	NewNodeInformerForKluster(clientFactory *kubernetes.SharedClientFactory, kluster *v1.Kluster, stopCh chan struct{}) (*NodeInformer, error)
	Run()
	Close()

	AddEventHandlerAddFunc(addFunc func(obj interface{}))
	AddEventHandlerUpdateFunc(updateFunc func(cur, old interface{}))
	AddEventHandlerDeleteFunc(deleteFunc func(obj interface{}))

	GetKluster() *v1.Kluster
	GetNodeByKey(key string) (*api_v1.Node, error)
}
