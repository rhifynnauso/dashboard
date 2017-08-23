// Copyright 2017 The Kubernetes Dashboard Authors.
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

package pod

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	errorHandler "github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/controller"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"k8s.io/apimachinery/pkg/api/errors"
	res "k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

// PodDetail is a presentation layer view of Kubernetes PodDetail resource. This means it is
// PodDetail plus additional augmented data we can get from other sources (like services that
// target it).
type PodDetail struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Status of the Pod. See Kubernetes API for reference.
	PodPhase v1.PodPhase `json:"podPhase"`

	// IP address of the Pod.
	PodIP string `json:"podIP"`

	// Name of the Node this Pod runs on.
	NodeName string `json:"nodeName"`

	// Count of containers restarts.
	RestartCount int32 `json:"restartCount"`

	// Reference to the Controller
	Controller controller.ResourceOwner `json:"controller"`

	// List of container of this pod.
	Containers []Container `json:"containers"`

	// List of initContainer of this pod.
	InitContainers []Container `json:"initContainers"`

	// Metrics collected for this resource
	Metrics []metricapi.Metric `json:"metrics"`

	// Conditions of this pod.
	Conditions []common.Condition `json:"conditions"`

	// Events is list of events associated with a pod.
	EventList common.EventList `json:"eventList"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// Container represents a docker/rkt/etc. container that lives in a pod.
type Container struct {
	// Name of the container.
	Name string `json:"name"`

	// Image URI of the container.
	Image string `json:"image"`

	// List of environment variables.
	Env []EnvVar `json:"env"`

	// Commands of the container
	Commands []string `json:"commands"`

	// Command arguments
	Args []string `json:"args"`
}

// EnvVar represents an environment variable of a container.
type EnvVar struct {
	// Name of the variable.
	Name string `json:"name"`

	// Value of the variable. May be empty if value from is defined.
	Value string `json:"value"`

	// Defined for derived variables. If non-null, the value is get from the reference.
	// Note that this is an API struct. This is intentional, as EnvVarSources are plain struct
	// references.
	ValueFrom *v1.EnvVarSource `json:"valueFrom"`
}

// GetPodDetail returns the details (PodDetail) of a named Pod from a particular namespace.
// TODO(maciaszczykm): Owner reference should be used instead of created by annotation.
func GetPodDetail(client kubernetes.Interface, metricClient metricapi.MetricClient, namespace, name string) (*PodDetail, error) {
	log.Printf("Getting details of %s pod in %s namespace", name, namespace)

	channels := &common.ResourceChannels{
		ConfigMapList: common.GetConfigMapListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
		SecretList:    common.GetSecretListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
	}

	pod, err := client.CoreV1().Pods(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	controller, err := getPodController(client, common.NewSameNamespaceQuery(namespace), pod)
	if err != nil {
		return nil, err
	}

	_, metricPromises := dataselect.GenericDataSelectWithMetrics(toCells([]v1.Pod{*pod}),
		dataselect.StdMetricsDataSelect, metricapi.NoResourceCache, metricClient)
	metrics, _ := metricPromises.GetMetrics()

	configMapList := <-channels.ConfigMapList.List
	err = <-channels.ConfigMapList.Error
	nonCriticalErrors, criticalError := errorHandler.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	secretList := <-channels.SecretList.List
	err = <-channels.SecretList.Error
	nonCriticalErrors, criticalError = errorHandler.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	eventList, err := GetEventsForPod(client, dataselect.DefaultDataSelect, pod.Namespace, pod.Name)
	nonCriticalErrors, criticalError = errorHandler.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	podDetail := toPodDetail(pod, metrics, configMapList, secretList, controller, eventList, nonCriticalErrors)
	return &podDetail, nil
}

func getPodController(client kubernetes.Interface, nsQuery *common.NamespaceQuery, pod *v1.Pod) (
	ctrl controller.ResourceOwner, err error) {

	channels := &common.ResourceChannels{
		PodList:   common.GetPodListChannel(client, nsQuery, 1),
		EventList: common.GetEventListChannel(client, nsQuery, 1),
	}

	pods := <-channels.PodList.List
	err = <-channels.PodList.Error
	if err != nil {
		return
	}

	events := <-channels.EventList.List
	if err := <-channels.EventList.Error; err != nil {
		events = &v1.EventList{}
	}

	ownerRef := metaV1.GetControllerOf(pod)
	if ownerRef != nil {
		var rc controller.ResourceController
		rc, err = controller.NewResourceController(*ownerRef, pod.Namespace, client)
		if err == nil {
			ctrl = rc.Get(pods.Items, events.Items)
		}
	}

	return
}

// isNotFoundError returns true when the given error is 404-NotFound error.
func isNotFoundError(err error) bool {
	statusErr, ok := err.(*errors.StatusError)
	if !ok {
		return false
	}
	return statusErr.ErrStatus.Code == 404
}

func extractContainerInfo(containerList []v1.Container, pod *v1.Pod, configMaps *v1.ConfigMapList, secrets *v1.SecretList) []Container {
	containers := make([]Container, 0)
	for _, container := range containerList {
		vars := make([]EnvVar, 0)
		for _, envVar := range container.Env {
			variable := EnvVar{
				Name:      envVar.Name,
				Value:     envVar.Value,
				ValueFrom: envVar.ValueFrom,
			}
			if variable.ValueFrom != nil {
				variable.Value = evalValueFrom(variable.ValueFrom, &container, pod,
					configMaps, secrets)
			}
			vars = append(vars, variable)
		}
		containers = append(containers, Container{
			Name:     container.Name,
			Image:    container.Image,
			Env:      vars,
			Commands: container.Command,
			Args:     container.Args,
		})
	}
	return containers
}

func toPodDetail(pod *v1.Pod, metrics []metricapi.Metric, configMaps *v1.ConfigMapList, secrets *v1.SecretList,
	controller controller.ResourceOwner, events *common.EventList, nonCriticalErrors []error) PodDetail {
	return PodDetail{
		ObjectMeta:     api.NewObjectMeta(pod.ObjectMeta),
		TypeMeta:       api.NewTypeMeta(api.ResourceKindPod),
		PodPhase:       pod.Status.Phase,
		PodIP:          pod.Status.PodIP,
		RestartCount:   getRestartCount(*pod),
		NodeName:       pod.Spec.NodeName,
		Controller:     controller,
		Containers:     extractContainerInfo(pod.Spec.Containers, pod, configMaps, secrets),
		InitContainers: extractContainerInfo(pod.Spec.InitContainers, pod, configMaps, secrets),
		Metrics:        metrics,
		Conditions:     getPodConditions(*pod),
		EventList:      *events,
		Errors:         nonCriticalErrors,
	}
}

// evalValueFrom evaluates environment value from given source. For more details check:
// https://github.com/kubernetes/kubernetes/blob/d82e51edc5f02bff39661203c9b503d054c3493b/pkg/kubectl/describe.go#L1056
func evalValueFrom(src *v1.EnvVarSource, container *v1.Container, pod *v1.Pod,
	configMaps *v1.ConfigMapList, secrets *v1.SecretList) string {
	switch {
	case src.ConfigMapKeyRef != nil:
		name := src.ConfigMapKeyRef.LocalObjectReference.Name
		for _, configMap := range configMaps.Items {
			if configMap.ObjectMeta.Name == name {
				return configMap.Data[src.ConfigMapKeyRef.Key]
			}
		}
	case src.SecretKeyRef != nil:
		name := src.SecretKeyRef.LocalObjectReference.Name
		for _, secret := range secrets.Items {
			if secret.ObjectMeta.Name == name {
				return base64.StdEncoding.EncodeToString([]byte(
					secret.Data[src.SecretKeyRef.Key]))
			}
		}
	case src.ResourceFieldRef != nil:
		valueFrom, err := extractContainerResourceValue(src.ResourceFieldRef, container)
		if err != nil {
			valueFrom = ""
		}
		resource := src.ResourceFieldRef.Resource
		if valueFrom == "0" && (resource == "limits.cpu" || resource == "limits.memory") {
			valueFrom = "node allocatable"
		}
		return valueFrom
	case src.FieldRef != nil:
		internalFieldPath, _, err := kubeapi.Scheme.ConvertFieldLabel(src.FieldRef.APIVersion,
			"Pod", src.FieldRef.FieldPath, "")
		if err != nil {
			log.Println(err)
			return ""
		}
		valueFrom, err := ExtractFieldPathAsString(pod, internalFieldPath)
		if err != nil {
			log.Println(err)
			return ""
		}
		return valueFrom
	}
	return ""
}

// extractContainerResourceValue extracts the value of a resource in an already known container.
func extractContainerResourceValue(fs *v1.ResourceFieldSelector, container *v1.Container) (string,
	error) {
	divisor := res.Quantity{}
	if divisor.Cmp(fs.Divisor) == 0 {
		divisor = res.MustParse("1")
	} else {
		divisor = fs.Divisor
	}

	switch fs.Resource {
	case "limits.cpu":
		return strconv.FormatInt(int64(math.Ceil(float64(container.Resources.Limits.
			Cpu().MilliValue())/float64(divisor.MilliValue()))), 10), nil
	case "limits.memory":
		return strconv.FormatInt(int64(math.Ceil(float64(container.Resources.Limits.
			Memory().Value())/float64(divisor.Value()))), 10), nil
	case "requests.cpu":
		return strconv.FormatInt(int64(math.Ceil(float64(container.Resources.Requests.
			Cpu().MilliValue())/float64(divisor.MilliValue()))), 10), nil
	case "requests.memory":
		return strconv.FormatInt(int64(math.Ceil(float64(container.Resources.Requests.
			Memory().Value())/float64(divisor.Value()))), 10), nil
	}

	return "", fmt.Errorf("Unsupported container resource : %v", fs.Resource)
}
