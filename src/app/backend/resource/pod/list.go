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

package pod

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/integration/metric/heapster"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"
	"github.com/kubernetes/dashboard/src/app/backend/resource/metric"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// PodList contains a list of Pods in the cluster.
type PodList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	// Unordered list of Pods.
	Pods              []Pod           `json:"pods"`
	CumulativeMetrics []metric.Metric `json:"cumulativeMetrics"`
}

type PodStatus struct {
	// Status of the Pod. See Kubernetes API for reference.
	Status          string              `json:"status"`
	PodPhase        v1.PodPhase         `json:"podPhase"`
	ContainerStates []v1.ContainerState `json:"containerStates"`
}

// Pod is a presentation layer view of Kubernetes Pod resource. This means
// it is Pod plus additional augmented data we can get from other sources
// (like services that target it).
type Pod struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// More info on pod status
	PodStatus PodStatus `json:"podStatus"`

	// Count of containers restarts.
	RestartCount int32 `json:"restartCount"`

	// Pod metrics.
	Metrics *common.PodMetrics `json:"metrics"`

	// Pod warning events
	Warnings []common.Event `json:"warnings"`
}

// GetPodList returns a list of all Pods in the cluster.
func GetPodList(client k8sClient.Interface, heapsterClient heapster.HeapsterClient,
	nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*PodList, error) {
	log.Print("Getting list of all pods in the cluster")

	channels := &common.ResourceChannels{
		PodList:   common.GetPodListChannelWithOptions(client, nsQuery, metaV1.ListOptions{}, 1),
		EventList: common.GetEventListChannel(client, nsQuery, 1),
	}

	return GetPodListFromChannels(channels, dsQuery, heapsterClient)
}

// GetPodListFromChannels returns a list of all Pods in the cluster
// reading required resource list once from the channels.
func GetPodListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery,
	heapsterClient heapster.HeapsterClient) (*PodList, error) {

	pods := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	eventList := <-channels.EventList.List
	if err := <-channels.EventList.Error; err != nil {
		return nil, err
	}

	podList := CreatePodList(pods.Items, eventList.Items, dsQuery, heapsterClient)
	return &podList, nil
}

func CreatePodList(pods []v1.Pod, events []v1.Event, dsQuery *dataselect.DataSelectQuery,
	heapsterClient heapster.HeapsterClient) PodList {

	channels := &common.ResourceChannels{
		PodMetrics: common.GetPodListMetricsChannel(heapsterClient, pods, 1),
	}

	if err := <-channels.PodMetrics.Error; err != nil {
		log.Printf("Skipping Heapster metrics because of error: %s\n", err)
	}
	metrics := <-channels.PodMetrics.MetricsByPod

	podList := PodList{
		Pods: make([]Pod, 0),
	}

	cache := &dataselect.CachedResources{Pods: pods}

	podCells, cumulativeMetricsPromises, filteredTotal := dataselect.GenericDataSelectWithFilterAndMetrics(toCells(pods), dsQuery,
		cache, &heapsterClient)
	pods = fromCells(podCells)
	podList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, pod := range pods {
		warnings := event.GetPodsEventWarnings(events, []v1.Pod{pod})

		podDetail := ToPod(&pod, metrics, warnings)
		podDetail.Warnings = warnings
		podList.Pods = append(podList.Pods, podDetail)

	}
	cumulativeMetrics, err := cumulativeMetricsPromises.GetMetrics()

	podList.CumulativeMetrics = cumulativeMetrics
	if err != nil {
		podList.CumulativeMetrics = make([]metric.Metric, 0)
	}

	return podList
}
