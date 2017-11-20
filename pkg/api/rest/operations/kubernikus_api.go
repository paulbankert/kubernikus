// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"
	"strings"

	errors "github.com/go-openapi/errors"
	loads "github.com/go-openapi/loads"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	security "github.com/go-openapi/runtime/security"
	spec "github.com/go-openapi/spec"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/sapcc/kubernikus/pkg/api/models"
)

// NewKubernikusAPI creates a new Kubernikus instance
func NewKubernikusAPI(spec *loads.Document) *KubernikusAPI {
	return &KubernikusAPI{
		handlers:            make(map[string]map[string]http.Handler),
		formats:             strfmt.Default,
		defaultConsumes:     "application/json",
		defaultProduces:     "application/json",
		ServerShutdown:      func() {},
		spec:                spec,
		ServeError:          errors.ServeError,
		BasicAuthenticator:  security.BasicAuth,
		APIKeyAuthenticator: security.APIKeyAuth,
		BearerAuthenticator: security.BearerAuth,
		JSONConsumer:        runtime.JSONConsumer(),
		JSONProducer:        runtime.JSONProducer(),
		CreateClusterHandler: CreateClusterHandlerFunc(func(params CreateClusterParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation CreateCluster has not yet been implemented")
		}),
		GetClusterCredentialsHandler: GetClusterCredentialsHandlerFunc(func(params GetClusterCredentialsParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation GetClusterCredentials has not yet been implemented")
		}),
		GetClusterEventsHandler: GetClusterEventsHandlerFunc(func(params GetClusterEventsParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation GetClusterEvents has not yet been implemented")
		}),
		GetClusterInfoHandler: GetClusterInfoHandlerFunc(func(params GetClusterInfoParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation GetClusterInfo has not yet been implemented")
		}),
		GetOpenstackMetadataHandler: GetOpenstackMetadataHandlerFunc(func(params GetOpenstackMetadataParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation GetOpenstackMetadata has not yet been implemented")
		}),
		InfoHandler: InfoHandlerFunc(func(params InfoParams) middleware.Responder {
			return middleware.NotImplemented("operation Info has not yet been implemented")
		}),
		ListAPIVersionsHandler: ListAPIVersionsHandlerFunc(func(params ListAPIVersionsParams) middleware.Responder {
			return middleware.NotImplemented("operation ListAPIVersions has not yet been implemented")
		}),
		ListClustersHandler: ListClustersHandlerFunc(func(params ListClustersParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation ListClusters has not yet been implemented")
		}),
		ShowClusterHandler: ShowClusterHandlerFunc(func(params ShowClusterParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation ShowCluster has not yet been implemented")
		}),
		TerminateClusterHandler: TerminateClusterHandlerFunc(func(params TerminateClusterParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation TerminateCluster has not yet been implemented")
		}),
		UpdateClusterHandler: UpdateClusterHandlerFunc(func(params UpdateClusterParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation UpdateCluster has not yet been implemented")
		}),

		// Applies when the "x-auth-token" header is set
		KeystoneAuth: func(token string) (*models.Principal, error) {
			return nil, errors.NotImplemented("api key auth (keystone) x-auth-token from header param [x-auth-token] has not yet been implemented")
		},

		// default authorizer is authorized meaning no requests are blocked
		APIAuthorizer: security.Authorized(),
	}
}

/*KubernikusAPI the kubernikus API */
type KubernikusAPI struct {
	spec            *loads.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	defaultConsumes string
	defaultProduces string
	Middleware      func(middleware.Builder) http.Handler

	// BasicAuthenticator generates a runtime.Authenticator from the supplied basic auth function.
	// It has a default implemention in the security package, however you can replace it for your particular usage.
	BasicAuthenticator func(security.UserPassAuthentication) runtime.Authenticator
	// APIKeyAuthenticator generates a runtime.Authenticator from the supplied token auth function.
	// It has a default implemention in the security package, however you can replace it for your particular usage.
	APIKeyAuthenticator func(string, string, security.TokenAuthentication) runtime.Authenticator
	// BearerAuthenticator generates a runtime.Authenticator from the supplied bearer token auth function.
	// It has a default implemention in the security package, however you can replace it for your particular usage.
	BearerAuthenticator func(string, security.ScopedTokenAuthentication) runtime.Authenticator

	// JSONConsumer registers a consumer for a "application/json" mime type
	JSONConsumer runtime.Consumer

	// JSONProducer registers a producer for a "application/json" mime type
	JSONProducer runtime.Producer

	// KeystoneAuth registers a function that takes a token and returns a principal
	// it performs authentication based on an api key x-auth-token provided in the header
	KeystoneAuth func(string) (*models.Principal, error)

	// APIAuthorizer provides access control (ACL/RBAC/ABAC) by providing access to the request and authenticated principal
	APIAuthorizer runtime.Authorizer

	// CreateClusterHandler sets the operation handler for the create cluster operation
	CreateClusterHandler CreateClusterHandler
	// GetClusterCredentialsHandler sets the operation handler for the get cluster credentials operation
	GetClusterCredentialsHandler GetClusterCredentialsHandler
	// GetClusterEventsHandler sets the operation handler for the get cluster events operation
	GetClusterEventsHandler GetClusterEventsHandler
	// GetClusterInfoHandler sets the operation handler for the get cluster info operation
	GetClusterInfoHandler GetClusterInfoHandler
	// GetOpenstackMetadataHandler sets the operation handler for the get openstack metadata operation
	GetOpenstackMetadataHandler GetOpenstackMetadataHandler
	// InfoHandler sets the operation handler for the info operation
	InfoHandler InfoHandler
	// ListAPIVersionsHandler sets the operation handler for the list API versions operation
	ListAPIVersionsHandler ListAPIVersionsHandler
	// ListClustersHandler sets the operation handler for the list clusters operation
	ListClustersHandler ListClustersHandler
	// ShowClusterHandler sets the operation handler for the show cluster operation
	ShowClusterHandler ShowClusterHandler
	// TerminateClusterHandler sets the operation handler for the terminate cluster operation
	TerminateClusterHandler TerminateClusterHandler
	// UpdateClusterHandler sets the operation handler for the update cluster operation
	UpdateClusterHandler UpdateClusterHandler

	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup

	// User defined logger function.
	Logger func(string, ...interface{})
}

// SetDefaultProduces sets the default produces media type
func (o *KubernikusAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *KubernikusAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// SetSpec sets a spec that will be served for the clients.
func (o *KubernikusAPI) SetSpec(spec *loads.Document) {
	o.spec = spec
}

// DefaultProduces returns the default produces media type
func (o *KubernikusAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *KubernikusAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *KubernikusAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *KubernikusAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the KubernikusAPI
func (o *KubernikusAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}

	if o.KeystoneAuth == nil {
		unregistered = append(unregistered, "XAuthTokenAuth")
	}

	if o.CreateClusterHandler == nil {
		unregistered = append(unregistered, "CreateClusterHandler")
	}

	if o.GetClusterCredentialsHandler == nil {
		unregistered = append(unregistered, "GetClusterCredentialsHandler")
	}

	if o.GetClusterEventsHandler == nil {
		unregistered = append(unregistered, "GetClusterEventsHandler")
	}

	if o.GetClusterInfoHandler == nil {
		unregistered = append(unregistered, "GetClusterInfoHandler")
	}

	if o.GetOpenstackMetadataHandler == nil {
		unregistered = append(unregistered, "GetOpenstackMetadataHandler")
	}

	if o.InfoHandler == nil {
		unregistered = append(unregistered, "InfoHandler")
	}

	if o.ListAPIVersionsHandler == nil {
		unregistered = append(unregistered, "ListAPIVersionsHandler")
	}

	if o.ListClustersHandler == nil {
		unregistered = append(unregistered, "ListClustersHandler")
	}

	if o.ShowClusterHandler == nil {
		unregistered = append(unregistered, "ShowClusterHandler")
	}

	if o.TerminateClusterHandler == nil {
		unregistered = append(unregistered, "TerminateClusterHandler")
	}

	if o.UpdateClusterHandler == nil {
		unregistered = append(unregistered, "UpdateClusterHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *KubernikusAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *KubernikusAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]runtime.Authenticator {

	result := make(map[string]runtime.Authenticator)
	for name, scheme := range schemes {
		switch name {

		case "keystone":

			result[name] = o.APIKeyAuthenticator(scheme.Name, scheme.In, func(token string) (interface{}, error) {
				return o.KeystoneAuth(token)
			})

		}
	}
	return result

}

// Authorizer returns the registered authorizer
func (o *KubernikusAPI) Authorizer() runtime.Authorizer {

	return o.APIAuthorizer

}

// ConsumersFor gets the consumers for the specified media types
func (o *KubernikusAPI) ConsumersFor(mediaTypes []string) map[string]runtime.Consumer {

	result := make(map[string]runtime.Consumer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONConsumer

		}
	}
	return result

}

// ProducersFor gets the producers for the specified media types
func (o *KubernikusAPI) ProducersFor(mediaTypes []string) map[string]runtime.Producer {

	result := make(map[string]runtime.Producer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONProducer

		}
	}
	return result

}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *KubernikusAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	if path == "/" {
		path = ""
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

// Context returns the middleware context for the kubernikus API
func (o *KubernikusAPI) Context() *middleware.Context {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	return o.context
}

func (o *KubernikusAPI) initHandlerCache() {
	o.Context() // don't care about the result, just that the initialization happened

	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/api/v1/clusters"] = NewCreateCluster(o.context, o.CreateClusterHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/clusters/{name}/credentials"] = NewGetClusterCredentials(o.context, o.GetClusterCredentialsHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/clusters/{name}/events"] = NewGetClusterEvents(o.context, o.GetClusterEventsHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/clusters/{name}/info"] = NewGetClusterInfo(o.context, o.GetClusterInfoHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/openstack/metadata"] = NewGetOpenstackMetadata(o.context, o.GetOpenstackMetadataHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/info"] = NewInfo(o.context, o.InfoHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api"] = NewListAPIVersions(o.context, o.ListAPIVersionsHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/clusters"] = NewListClusters(o.context, o.ListClustersHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/clusters/{name}"] = NewShowCluster(o.context, o.ShowClusterHandler)

	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/api/v1/clusters/{name}"] = NewTerminateCluster(o.context, o.TerminateClusterHandler)

	if o.handlers["PUT"] == nil {
		o.handlers["PUT"] = make(map[string]http.Handler)
	}
	o.handlers["PUT"]["/api/v1/clusters/{name}"] = NewUpdateCluster(o.context, o.UpdateClusterHandler)

}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *KubernikusAPI) Serve(builder middleware.Builder) http.Handler {
	o.Init()

	if o.Middleware != nil {
		return o.Middleware(builder)
	}
	return o.context.APIHandler(builder)
}

// Init allows you to just initialize the handler cache, you can then recompose the middelware as you see fit
func (o *KubernikusAPI) Init() {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}
}
