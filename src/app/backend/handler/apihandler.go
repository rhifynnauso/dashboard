// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	restful "github.com/emicklei/go-restful"
	// TODO(maciaszczykm): Avoid using dot-imports.
	. "github.com/kubernetes/dashboard/client"
	"github.com/kubernetes/dashboard/resource/common"
	. "github.com/kubernetes/dashboard/resource/container"
	"github.com/kubernetes/dashboard/resource/deployment"
	. "github.com/kubernetes/dashboard/resource/namespace"
	"github.com/kubernetes/dashboard/resource/pod"
	"github.com/kubernetes/dashboard/resource/replicaset"
	. "github.com/kubernetes/dashboard/resource/replicationcontroller"
	. "github.com/kubernetes/dashboard/resource/secret"
	resourceService "github.com/kubernetes/dashboard/resource/service"
	"github.com/kubernetes/dashboard/resource/workload"
	. "github.com/kubernetes/dashboard/validation"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

const (
	// RequestLogString is a template for request log message.
	RequestLogString = "Incoming %s %s %s request from %s"

	// ResponseLogString is a template for response log message.
	ResponseLogString = "Outcoming response to %s with %d status code"
)

// ApiHandler is a representation of API handler. Structure contains client, Heapster client and
// client configuration.
type ApiHandler struct {
	client         *client.Client
	heapsterClient HeapsterClient
	clientConfig   clientcmd.ClientConfig
	verber         common.ResourceVerber
}

// Web-service filter function used for request and response logging.
func wsLogger(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	log.Printf(FormatRequestLog(req))
	chain.ProcessFilter(req, resp)
	log.Printf(FormatResponseLog(resp, req))
}

// FormatRequestLog formats request log string.
// TODO(maciaszczykm): Display request body.
func FormatRequestLog(req *restful.Request) string {
	reqURI := ""
	if req.Request.URL != nil {
		reqURI = req.Request.URL.RequestURI()
	}

	return fmt.Sprintf(RequestLogString, req.Request.Proto, req.Request.Method,
		reqURI, req.Request.RemoteAddr)
}

// FormatResponseLog formats response log string.
// TODO(maciaszczykm): Display response content.
func FormatResponseLog(resp *restful.Response, req *restful.Request) string {
	return fmt.Sprintf(ResponseLogString, req.Request.RemoteAddr, resp.StatusCode())
}

// CreateHttpApiHandler creates a new HTTP handler that handles all requests to the API of the backend.
func CreateHttpApiHandler(client *client.Client, heapsterClient HeapsterClient,
	clientConfig clientcmd.ClientConfig) http.Handler {

	verber := common.NewResourceVerber(client.RESTClient)
	apiHandler := ApiHandler{client, heapsterClient, clientConfig, verber}
	wsContainer := restful.NewContainer()

	deployWs := new(restful.WebService)
	deployWs.Filter(wsLogger)
	deployWs.Path("/api/v1/appdeployments").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	deployWs.Route(
		deployWs.POST("").
			To(apiHandler.handleDeploy).
			Reads(AppDeploymentSpec{}).
			Writes(AppDeploymentSpec{}))
	deployWs.Route(
		deployWs.POST("/validate/name").
			To(apiHandler.handleNameValidity).
			Reads(AppNameValiditySpec{}).
			Writes(AppNameValidity{}))
	deployWs.Route(
		deployWs.POST("/validate/imagereference").
			To(apiHandler.handleImageReferenceValidity).
			Reads(ImageReferenceValiditySpec{}).
			Writes(ImageReferenceValidity{}))
	deployWs.Route(
		deployWs.POST("/validate/protocol").
			To(apiHandler.handleProtocolValidity).
			Reads(ProtocolValiditySpec{}).
			Writes(ProtocolValidity{}))
	deployWs.Route(
		deployWs.GET("/protocols").
			To(apiHandler.handleGetAvailableProcotols).
			Writes(Protocols{}))
	wsContainer.Add(deployWs)

	deployFromFileWs := new(restful.WebService)
	deployFromFileWs.Path("/api/v1/appdeploymentfromfile").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	deployFromFileWs.Route(
		deployFromFileWs.POST("").
			To(apiHandler.handleDeployFromFile).
			Reads(AppDeploymentFromFileSpec{}).
			Writes(AppDeploymentFromFileResponse{}))
	wsContainer.Add(deployFromFileWs)

	replicationControllerWs := new(restful.WebService)
	replicationControllerWs.Filter(wsLogger)
	replicationControllerWs.Path("/api/v1/replicationcontrollers").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	replicationControllerWs.Route(
		replicationControllerWs.GET("").
			To(apiHandler.handleGetReplicationControllerList).
			Writes(ReplicationControllerList{}))
	replicationControllerWs.Route(
		replicationControllerWs.GET("/{namespace}/{replicationController}").
			To(apiHandler.handleGetReplicationControllerDetail).
			Writes(ReplicationControllerDetail{}))
	replicationControllerWs.Route(
		replicationControllerWs.POST("/{namespace}/{replicationController}/update/pods").
			To(apiHandler.handleUpdateReplicasCount).
			Reads(ReplicationControllerSpec{}))
	replicationControllerWs.Route(
		replicationControllerWs.DELETE("/{namespace}/{replicationController}").
			To(apiHandler.handleDeleteReplicationController))
	replicationControllerWs.Route(
		replicationControllerWs.GET("/pods/{namespace}/{replicationController}").
			To(apiHandler.handleGetReplicationControllerPods).
			Writes(ReplicationControllerPods{}))
	wsContainer.Add(replicationControllerWs)

	workloadsWs := new(restful.WebService)
	workloadsWs.Filter(wsLogger)
	workloadsWs.Path("/api/v1/workloads").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	workloadsWs.Route(
		workloadsWs.GET("").
			To(apiHandler.handleGetWorkloads).
			Writes(workload.Workloads{}))
	wsContainer.Add(workloadsWs)

	replicaSetsWs := new(restful.WebService)
	replicaSetsWs.Filter(wsLogger)
	replicaSetsWs.Path("/api/v1/replicasets").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	replicaSetsWs.Route(
		replicaSetsWs.GET("").
			To(apiHandler.handleGetReplicaSets).
			Writes(replicaset.ReplicaSetList{}))
	replicaSetsWs.Route(
		replicaSetsWs.GET("/{namespace}/{replicaSet}").
			To(apiHandler.handleGetReplicaSetDetail).
			Writes(replicaset.ReplicaSetDetail{}))
	wsContainer.Add(replicaSetsWs)

	podsWs := new(restful.WebService)
	podsWs.Filter(wsLogger)
	podsWs.Path("/api/v1/pods").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	podsWs.Route(
		podsWs.GET("").
			To(apiHandler.handleGetPods).
			Writes(pod.PodList{}))
	podsWs.Route(
		podsWs.GET("/{namespace}/{pod}").
		To(apiHandler.handleGetPodDetail).
		Writes(pod.PodDetail{}))
	wsContainer.Add(podsWs)

	deploymentsWs := new(restful.WebService)
	deploymentsWs.Filter(wsLogger)
	deploymentsWs.Path("/api/v1/deployments").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	deploymentsWs.Route(
		deploymentsWs.GET("").
			To(apiHandler.handleGetDeployments).
			Writes(deployment.DeploymentList{}))
	wsContainer.Add(deploymentsWs)

	namespacesWs := new(restful.WebService)
	namespacesWs.Filter(wsLogger)
	namespacesWs.Path("/api/v1/namespaces").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	namespacesWs.Route(
		namespacesWs.POST("").
			To(apiHandler.handleCreateNamespace).
			Reads(NamespaceSpec{}).
			Writes(NamespaceSpec{}))
	namespacesWs.Route(
		namespacesWs.GET("").
			To(apiHandler.handleGetNamespaces).
			Writes(NamespaceList{}))
	wsContainer.Add(namespacesWs)

	logsWs := new(restful.WebService)
	logsWs.Filter(wsLogger)
	logsWs.Path("/api/v1/logs").
		Produces(restful.MIME_JSON)
	logsWs.Route(
		logsWs.GET("/{namespace}/{podId}").
			To(apiHandler.handleLogs).
			Writes(Logs{}))
	logsWs.Route(
		logsWs.GET("/{namespace}/{podId}/{container}").
			To(apiHandler.handleLogs).
			Writes(Logs{}))
	wsContainer.Add(logsWs)

	eventsWs := new(restful.WebService)
	eventsWs.Filter(wsLogger)
	eventsWs.Path("/api/v1/events").
		Produces(restful.MIME_JSON)
	eventsWs.Route(
		eventsWs.GET("/{namespace}/{replicationController}").
			To(apiHandler.handleEvents).
			Writes(common.EventList{}))
	wsContainer.Add(eventsWs)

	secretsWs := new(restful.WebService)
	secretsWs.Path("/api/v1/secrets").Produces(restful.MIME_JSON)
	secretsWs.Route(
		secretsWs.GET("/{namespace}").
			To(apiHandler.handleGetSecrets).
			Writes(SecretsList{}))
	secretsWs.Route(
		secretsWs.POST("").
			To(apiHandler.handleCreateImagePullSecret).
			Reads(ImagePullSecretSpec{}).
			Writes(Secret{}))
	wsContainer.Add(secretsWs)

	servicesWs := new(restful.WebService)
	servicesWs.Filter(wsLogger)
	servicesWs.Path("/api/v1/services").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	servicesWs.Route(
		servicesWs.GET("").
			To(apiHandler.handleGetServiceList).
			Writes(resourceService.ServiceList{}))
	servicesWs.Route(
		servicesWs.GET("/{namespace}/{service}").
			To(apiHandler.handleGetServiceDetail).
			Writes(resourceService.ServiceDetail{}))
	wsContainer.Add(servicesWs)

	resourceVerberWs := new(restful.WebService)
	resourceVerberWs.Filter(wsLogger)
	resourceVerberWs.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	resourceVerberWs.Route(
		resourceVerberWs.DELETE("/{kind}/namespace/{namespace}/name/{name}").
			To(apiHandler.handleDeleteResource))
	wsContainer.Add(resourceVerberWs)

	return wsContainer
}

// Handles get service list API call.
func (apiHandler *ApiHandler) handleGetServiceList(request *restful.Request, response *restful.Response) {
	result, err := resourceService.GetServiceList(apiHandler.client)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get service detail API call.
func (apiHandler *ApiHandler) handleGetServiceDetail(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	result, err := resourceService.GetServiceDetail(apiHandler.client, apiHandler.heapsterClient,
		namespace, service)
	if err != nil {
		handleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles deploy API call.
func (apiHandler *ApiHandler) handleDeploy(request *restful.Request, response *restful.Response) {
	appDeploymentSpec := new(AppDeploymentSpec)
	if err := request.ReadEntity(appDeploymentSpec); err != nil {
		handleInternalError(response, err)
		return
	}
	if err := DeployApp(appDeploymentSpec, apiHandler.client); err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, appDeploymentSpec)
}

// Handles deploy from file API call.
func (apiHandler *ApiHandler) handleDeployFromFile(request *restful.Request, response *restful.Response) {
	deploymentSpec := new(AppDeploymentFromFileSpec)
	if err := request.ReadEntity(deploymentSpec); err != nil {
		handleInternalError(response, err)
		return
	}

	isDeployed, err := DeployAppFromFile(
		deploymentSpec, CreateObjectFromInfoFn, apiHandler.clientConfig)
	if !isDeployed {
		handleInternalError(response, err)
		return
	}

	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	response.WriteHeaderAndEntity(http.StatusCreated, AppDeploymentFromFileResponse{
		Name:    deploymentSpec.Name,
		Content: deploymentSpec.Content,
		Error:   errorMessage,
	})
}

// Handles app name validation API call.
func (apiHandler *ApiHandler) handleNameValidity(request *restful.Request, response *restful.Response) {
	spec := new(AppNameValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		handleInternalError(response, err)
		return
	}

	validity, err := ValidateAppName(spec, apiHandler.client)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, validity)
}

// Handles image reference validation API call.
func (ApiHandler *ApiHandler) handleImageReferenceValidity(request *restful.Request, response *restful.Response) {
	spec := new(ImageReferenceValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		handleInternalError(response, err)
		return
	}

	validity, err := ValidateImageReference(spec)
	if err != nil {
		handleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, validity)
}

// Handles protocol validation API call.
func (apiHandler *ApiHandler) handleProtocolValidity(request *restful.Request, response *restful.Response) {
	spec := new(ProtocolValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, ValidateProtocol(spec))
}

// Handles get available protocols API call.
func (apiHandler *ApiHandler) handleGetAvailableProcotols(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusCreated, GetAvailableProtocols())
}

// Handles get Replication Controller list API call.
func (apiHandler *ApiHandler) handleGetReplicationControllerList(
	request *restful.Request, response *restful.Response) {

	result, err := GetReplicationControllerList(apiHandler.client)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get Workloads list API call.
func (apiHandler *ApiHandler) handleGetWorkloads(
	request *restful.Request, response *restful.Response) {

	result, err := workload.GetWorkloads(apiHandler.client, apiHandler.heapsterClient)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get Replica Sets list API call.
func (apiHandler *ApiHandler) handleGetReplicaSets(
	request *restful.Request, response *restful.Response) {

	result, err := replicaset.GetReplicaSetList(apiHandler.client)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

func (apiHandler *ApiHandler) handleGetReplicaSetDetail(
	request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	result, err := replicaset.GetReplicaSetDetail(apiHandler.client, apiHandler.heapsterClient,
		namespace, replicaSet)

	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get Deployment list API call.
func (apiHandler *ApiHandler) handleGetDeployments(
	request *restful.Request, response *restful.Response) {

	result, err := deployment.GetDeploymentList(apiHandler.client)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get Pod list API call.
func (apiHandler *ApiHandler) handleGetPods(
	request *restful.Request, response *restful.Response) {

	result, err := pod.GetPodList(apiHandler.client, apiHandler.heapsterClient)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get Pod detail API call.
func (apiHandler *ApiHandler) handleGetPodDetail(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	result, err := pod.GetPodDetail(apiHandler.client, apiHandler.heapsterClient, namespace, podName)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles get Replication Controller detail API call.
func (apiHandler *ApiHandler) handleGetReplicationControllerDetail(
	request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	replicationController := request.PathParameter("replicationController")
	result, err := GetReplicationControllerDetail(apiHandler.client, apiHandler.heapsterClient, namespace, replicationController)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles update of Replication Controller pods update API call.
func (apiHandler *ApiHandler) handleUpdateReplicasCount(
	request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	replicationControllerName := request.PathParameter("replicationController")
	replicationControllerSpec := new(ReplicationControllerSpec)

	if err := request.ReadEntity(replicationControllerSpec); err != nil {
		handleInternalError(response, err)
		return
	}

	if err := UpdateReplicasCount(apiHandler.client, namespace, replicationControllerName,
		replicationControllerSpec); err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

// Handles delete Replication Controller API call.
// TODO(floreks): there has to be some kind of transaction here
func (apiHandler *ApiHandler) handleDeleteReplicationController(
	request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	replicationController := request.PathParameter("replicationController")
	deleteServices, err := strconv.ParseBool(request.QueryParameter("deleteServices"))
	if err != nil {
		handleInternalError(response, err)
		return
	}

	if err := DeleteReplicationController(apiHandler.client, namespace,
		replicationController, deleteServices); err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (apiHandler *ApiHandler) handleDeleteResource(
	request *restful.Request, response *restful.Response) {
	kind := request.PathParameter("kind")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	if err := apiHandler.verber.Delete(kind, namespace, name); err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

// Handles get Replication Controller Pods API call.
func (apiHandler *ApiHandler) handleGetReplicationControllerPods(
	request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	replicationController := request.PathParameter("replicationController")
	limit, err := strconv.Atoi(request.QueryParameter("limit"))
	if err != nil {
		limit = 0
	}
	result, err := GetReplicationControllerPods(apiHandler.client, namespace, replicationController, limit)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles namespace creation API call.
func (apiHandler *ApiHandler) handleCreateNamespace(request *restful.Request,
	response *restful.Response) {
	namespaceSpec := new(NamespaceSpec)
	if err := request.ReadEntity(namespaceSpec); err != nil {
		handleInternalError(response, err)
		return
	}
	if err := CreateNamespace(namespaceSpec, apiHandler.client); err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, namespaceSpec)
}

// Handles get namespace list API call.
func (apiHandler *ApiHandler) handleGetNamespaces(
	request *restful.Request, response *restful.Response) {

	result, err := GetNamespaceList(apiHandler.client)
	if err != nil {
		handleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles image pull secret creation API call.
func (apiHandler *ApiHandler) handleCreateImagePullSecret(request *restful.Request, response *restful.Response) {
	secretSpec := new(ImagePullSecretSpec)
	if err := request.ReadEntity(secretSpec); err != nil {
		handleInternalError(response, err)
		return
	}
	secret, err := CreateSecret(apiHandler.client, secretSpec)
	if err != nil {
		handleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, secret)
}

// Handles get secrets list API call.
func (apiHandler *ApiHandler) handleGetSecrets(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	result, err := GetSecrets(apiHandler.client, namespace)
	if err != nil {
		handleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// Handles log API call.
func (apiHandler *ApiHandler) handleLogs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	podId := request.PathParameter("podId")
	container := request.PathParameter("container")

	result, err := GetPodLogs(apiHandler.client, namespace, podId, container)
	if err != nil {
		handleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handles event API call.
func (apiHandler *ApiHandler) handleEvents(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	replicationController := request.PathParameter("replicationController")
	result, err := GetReplicationControllerEvents(apiHandler.client, namespace, replicationController)
	if err != nil {
		handleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, result)
}

// Handler that writes the given error to the response and sets appropriate HTTP status headers.
func handleInternalError(response *restful.Response, err error) {
	log.Print(err)
	response.AddHeader("Content-Type", "text/plain")
	response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
}
