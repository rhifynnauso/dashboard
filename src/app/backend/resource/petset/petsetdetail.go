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

package petset

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/apps"
	k8sClient "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/kubernetes/dashboard/src/app/backend/client"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
)

// PetSetDetail is a presentation layer view of Kubernetes Pet Set resource. This means
// it is Pet Set plus additional augmented data we can get from other sources
// (like services that target the same pods).
type PetSetDetail struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	// Aggregate information about pods belonging to this Pet Set.
	PodInfo common.PodInfo `json:"podInfo"`

	// Detailed information about Pods belonging to this Pet Set.
	PodList pod.PodList `json:"podList"`

	// Container images of the Pet Set.
	ContainerImages []string `json:"containerImages"`

	// List of events related to this Pet Set.
	EventList common.EventList `json:"eventList"`
}

// GetPetSetDetail gets pet set details.
func GetPetSetDetail(client *k8sClient.Client, heapsterClient client.HeapsterClient,
	namespace, name string) (*PetSetDetail, error) {

	log.Printf("Getting details of %s service in %s namespace", name, namespace)

	// TODO(floreks): Use channels.
	petSetData, err := client.Apps().PetSets(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
	}

	pods := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	events, err := GetPetSetEvents(client, petSetData.Namespace, petSetData.Name)
	if err != nil {
		return nil, err
	}

	petSet := getPetSetDetail(petSetData, heapsterClient, events, pods.Items)
	return &petSet, nil
}

func getPetSetDetail(petSet *apps.PetSet, heapsterClient client.HeapsterClient,
	events *common.EventList, pods []api.Pod) PetSetDetail {

	matchingPods := common.FilterNamespacedPodsByLabelSelector(pods, petSet.ObjectMeta.Namespace,
		petSet.Spec.Selector)

	podInfo := common.GetPodInfo(int32(petSet.Status.Replicas), int32(petSet.Spec.Replicas),
		matchingPods)

	return PetSetDetail{
		ObjectMeta:      common.NewObjectMeta(petSet.ObjectMeta),
		TypeMeta:        common.NewTypeMeta(common.ResourceKindPetSet),
		ContainerImages: common.GetContainerImages(&petSet.Spec.Template.Spec),
		PodInfo:         podInfo,
		// TODO(floreks): add pagination support
		PodList:   pod.CreatePodList(matchingPods, common.NO_PAGINATION, heapsterClient),
		EventList: *events,
	}
}
