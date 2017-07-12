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

package service

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// Service is a representation of a service.
type ServiceDetail struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// InternalEndpoint of all Kubernetes services that have the same label selector as connected Replication
	// Controller. Endpoint is DNS name merged with ports.
	InternalEndpoint common.Endpoint `json:"internalEndpoint"`

	// ExternalEndpoints of all Kubernetes services that have the same label selector as connected Replication
	// Controller. Endpoint is external IP address name merged with ports.
	ExternalEndpoints []common.Endpoint `json:"externalEndpoints"`

	// Label selector of the service.
	Selector map[string]string `json:"selector"`

	// Type determines how the service will be exposed.  Valid options: ClusterIP, NodePort, LoadBalancer
	Type v1.ServiceType `json:"type"`

	// ClusterIP is usually assigned by the master. Valid values are None, empty string (""), or
	// a valid IP address. None can be specified for headless services when proxying is not required
	ClusterIP string `json:"clusterIP"`

	// List of events related to this Service
	EventList common.EventList `json:"eventList"`

	// PodList represents list of pods targeted by same label selector as this service.
	PodList pod.PodList `json:"podList"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetServiceDetail gets service details.
func GetServiceDetail(client k8sClient.Interface, metricClient metricapi.MetricClient, namespace, name string,
	dsQuery *dataselect.DataSelectQuery) (*ServiceDetail, error) {

	log.Printf("Getting details of %s service in %s namespace", name, namespace)
	serviceData, err := client.CoreV1().Services(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	podList, err := GetServicePods(client, metricClient, namespace, name, dsQuery)
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	eventList, err := GetServiceEvents(client, dataselect.DefaultDataSelect, namespace, name)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	service := ToServiceDetail(serviceData)
	service.PodList = *podList
	service.EventList = *eventList
	service.Errors = nonCriticalErrors
	return &service, nil
}

// GetServicePods gets list of pods targeted by given label selector in given namespace.
func GetServicePods(client k8sClient.Interface, metricClient metricapi.MetricClient, namespace,
	name string, dsQuery *dataselect.DataSelectQuery) (*pod.PodList, error) {

	service, err := client.CoreV1().Services(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if service.Spec.Selector == nil {
		emptyPodList := &pod.PodList{
			Pods:              []pod.Pod{},
			CumulativeMetrics: []metricapi.Metric{},
		}
		return emptyPodList, nil
	}

	labelSelector := labels.SelectorFromSet(service.Spec.Selector)
	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannelWithOptions(client, common.NewSameNamespaceQuery(namespace),
			metaV1.ListOptions{
				LabelSelector: labelSelector.String(),
				FieldSelector: fields.Everything().String(),
			}, 1),
	}

	apiPodList := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	events, err := event.GetPodsEvents(client, namespace, apiPodList.Items)
	if err != nil {
		return nil, err
	}

	podList := pod.ToPodList(apiPodList.Items, events, []error{}, dsQuery, metricClient)
	return &podList, nil
}
