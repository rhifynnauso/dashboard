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

package replicationcontroller

import (
	"github.com/kubernetes/dashboard/src/app/backend/client"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"

	"k8s.io/kubernetes/pkg/api"
	k8sClient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"log"
)

// GetReplicationControllerPods return list of pods targeting replication controller associated
// to given name.
func GetReplicationControllerPods(client k8sClient.Interface, heapsterClient client.HeapsterClient,
	pQuery *common.PaginationQuery, rcName, namespace string) (*pod.PodList, error) {
	log.Printf("Getting replication controller %s pods in namespace %s", rcName, namespace)

	pods, err := getRawReplicationControllerPods(client, rcName, namespace)
	if err != nil {
		return nil, err
	}

	podList := pod.CreatePodList(pods, pQuery, heapsterClient)
	return &podList, nil
}

// Returns array of api pods targeting replication controller associated to given name.
func getRawReplicationControllerPods(client k8sClient.Interface, rcName, namespace string) (
	[]api.Pod, error) {

	replicationController, err := client.ReplicationControllers(namespace).Get(rcName)
	if err != nil {
		return nil, err
	}

	labelSelector := labels.SelectorFromSet(replicationController.Spec.Selector)
	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannelWithOptions(client, common.NewSameNamespaceQuery(namespace),
			api.ListOptions{
				LabelSelector: labelSelector,
				FieldSelector: fields.Everything(),
			}, 1),
	}

	podList := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	return podList.Items, nil
}

// Returns simple info about pods(running, desired, failing, etc.) related to given replication
// controller.
func getReplicationControllerPodInfo(client k8sClient.Interface, rc *api.ReplicationController,
	namespace string) (*common.PodInfo, error) {

	labelSelector := labels.SelectorFromSet(rc.Spec.Selector)
	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannelWithOptions(client, common.NewSameNamespaceQuery(namespace),
			api.ListOptions{
				LabelSelector: labelSelector,
				FieldSelector: fields.Everything(),
			}, 1),
	}

	pods := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	podInfo := common.GetPodInfo(rc.Status.Replicas, rc.Spec.Replicas, pods.Items)
	return &podInfo, nil
}
