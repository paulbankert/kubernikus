package kubernetes

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	kitlog "github.com/go-kit/kit/log"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiutilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	kubernikus_v1 "github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
)

type SharedClientFactory struct {
	clients          *sync.Map
	secretsInterface typedv1.SecretInterface
	Logger           kitlog.Logger
}

func NewSharedClientFactory(secrets typedv1.SecretInterface, klusterEvents cache.SharedIndexInformer, logger kitlog.Logger) *SharedClientFactory {
	factory := &SharedClientFactory{
		clients:          new(sync.Map),
		secretsInterface: secrets,
		Logger:           kitlog.With(logger, "client", "kubernetes"),
	}

	if klusterEvents != nil {
		klusterEvents.AddEventHandler(cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if kluster, ok := obj.(*kubernikus_v1.Kluster); ok {
					factory.clients.Delete(kluster.GetUID())
					factory.Logger.Log(
						"msg", "deleted shared kubernetes client",
						"kluster", kluster.GetName(),
						"project", kluster.Account(),
						"v", 2,
					)
				}
			},
		})
	}

	return factory
}

func (f *SharedClientFactory) ClientFor(k *kubernikus_v1.Kluster) (clientset kubernetes.Interface, err error) {
	defer func() {
		f.Logger.Log(
			"msg", "created shared kubernetes client",
			"kluster", k.GetName(),
			"project", k.Account(),
			"v", 2,
			"err", err,
		)
	}()

	if client, found := f.clients.Load(k.GetUID()); found {
		return client.(kubernetes.Interface), nil
	}
	secret, err := f.secretsInterface.Get(k.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	apiserverCACert, ok := secret.Data["tls-ca.pem"]
	if !ok {
		return nil, errors.New("tls-ca.pem not found in secret")
	}
	clientCert, ok := secret.Data["apiserver-clients-cluster-admin.pem"]
	if !ok {
		return nil, errors.New("tls-ca.pem not found in secret")
	}
	clientKey, ok := secret.Data["apiserver-clients-cluster-admin-key.pem"]
	if !ok {
		return nil, errors.New("tls-ca.pem not found in secret")
	}

	c := rest.Config{
		Host: k.Status.Apiserver,
		TLSClientConfig: rest.TLSClientConfig{
			CertData: clientCert,
			KeyData:  clientKey,
			CAData:   apiserverCACert,
		},
	}

	clientset, err = kubernetes.NewForConfig(&c)
	if err != nil {
		return nil, err
	}

	f.clients.Store(k.GetUID(), clientset)
	return clientset, nil

}

func NewConfig(kubeconfig, context string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	if len(context) > 0 {
		overrides.CurrentContext = context
	}

	if len(kubeconfig) > 0 {
		rules.ExplicitPath = kubeconfig
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}

func NewClient(kubeconfig, context string, logger kitlog.Logger) (kubernetes.Interface, error) {
	config, err := NewConfig(kubeconfig, context)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	logger.Log(
		"msg", "created new kubernetes client",
		"host", config.Host,
		"v", 3,
	)

	return clientset, nil
}

func NewClientConfigV1(name, user, url string, key, cert, ca []byte) clientcmdapiv1.Config {
	return clientcmdapiv1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: name,
		Clusters: []clientcmdapiv1.NamedCluster{
			clientcmdapiv1.NamedCluster{
				Name: name,
				Cluster: clientcmdapiv1.Cluster{
					Server: url,
					CertificateAuthorityData: ca,
				},
			},
		},
		Contexts: []clientcmdapiv1.NamedContext{
			clientcmdapiv1.NamedContext{
				Name: name,
				Context: clientcmdapiv1.Context{
					Cluster:  name,
					AuthInfo: user,
				},
			},
		},
		AuthInfos: []clientcmdapiv1.NamedAuthInfo{
			clientcmdapiv1.NamedAuthInfo{
				Name: user,
				AuthInfo: clientcmdapiv1.AuthInfo{
					ClientCertificateData: cert,
					ClientKeyData:         key,
				},
			},
		},
	}
}

func EnsureCRD(clientset apiextensionsclient.Interface, logger kitlog.Logger) error {
	klusterCRDName := kubernikus_v1.KlusterResourcePlural + "." + kubernikus_v1.GroupName
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: klusterCRDName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   kubernikus_v1.GroupName,
			Version: kubernikus_v1.SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: kubernikus_v1.KlusterResourcePlural,
				Kind:   reflect.TypeOf(kubernikus_v1.Kluster{}).Name(),
			},
		},
	}
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	//TODO: Should this error if it already exit?
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	// wait for CRD being established
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(klusterCRDName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					logger.Log(
						"msg", "name conflict while ensuring CRD",
						"reason", cond.Reason,
					)
				}
			}
		}
		return false, err
	})
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(klusterCRDName, nil)
		if deleteErr != nil {
			return apiutilerrors.NewAggregate([]error{err, deleteErr})
		}
		return err
	}
	return nil
}

func WaitForServer(client kubernetes.Interface, stopCh <-chan struct{}, logger kitlog.Logger) error {
	var healthzContent string

	err := wait.PollUntil(time.Second, func() (bool, error) {
		healthStatus := 0
		resp := client.Discovery().RESTClient().Get().AbsPath("/healthz").Do().StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			logger.Log(
				"msg", "server isn't health yet. Waiting a little while.",
			)
			return false, nil
		}
		content, _ := resp.Raw()
		healthzContent = string(content)

		return true, nil
	}, stopCh)
	if err != nil {
		return fmt.Errorf("Failed to contact apiserver. Last health: %v  Error: %v", healthzContent, err)
	}

	return nil
}
