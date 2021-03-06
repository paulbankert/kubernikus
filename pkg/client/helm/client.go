package helm

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-kit/kit/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"

	"github.com/sapcc/kubernikus/pkg/client/helm/portforwarder"
)

func NewClient(kubeClient kubernetes.Interface, kubeConfig *rest.Config, logger log.Logger) (*helm.Client, error) {
	logger = log.With(logger, "client", "helm")

	tillerHost := os.Getenv("TILLER_DEPLOY_SERVICE_HOST")
	if tillerHost == "" {
		tillerHost = "tiller-deploy.kube-system"
	}
	tillerPort := os.Getenv("TILLER_DEPLOY_SERVICE_PORT")
	if tillerPort == "" {
		tillerPort = "44134"
	}
	tillerHost = fmt.Sprintf("%s:%s", tillerHost, tillerPort)

	if _, err := rest.InClusterConfig(); err != nil {
		logger.Log(
			"msg", "We are not running inside the cluster. Creating tunnel to tiller pod.",
			"v", 2)
		tunnel, err := portforwarder.New("kube-system", kubeClient, kubeConfig)
		if err != nil {
			return nil, err
		}
		tillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)
		client := helm.NewClient(helm.Host(tillerHost))
		//Lets see how this goes: We close the tunnel as soon as the client is GC'ed.
		runtime.SetFinalizer(client, func(_ *helm.Client) {
			logger.Log(
				"msg", "trearing down tunnel to tiller",
				"host", tillerHost,
				"v", 2)
			tunnel.Close()
		})
		return client, nil
	}
	return helm.NewClient(helm.Host(tillerHost)), nil
}
