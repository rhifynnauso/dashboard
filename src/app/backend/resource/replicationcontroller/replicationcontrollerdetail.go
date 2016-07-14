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
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/client"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	resourceService "github.com/kubernetes/dashboard/src/app/backend/resource/service"

	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	"k8s.io/kubernetes/pkg/api"
	k8sClient "k8s.io/kubernetes/pkg/client/unversioned"
)

// ReplicationControllerDetail represents detailed information about a Replication Controller.
type ReplicationControllerDetail struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	// Label selector of the Replication Controller.
	LabelSelector map[string]string `json:"labelSelector"`

	// Container image list of the pod template specified by this Replication Controller.
	ContainerImages []string `json:"containerImages"`

	// Aggregate information about pods of this replication controller.
	PodInfo common.PodInfo `json:"podInfo"`

	// Detailed information about Pods belonging to this Replication Controller.
	PodList pod.PodList `json:"podList"`

	// Detailed information about service related to Replication Controller.
	ServiceList resourceService.ServiceList `json:"serviceList"`

	// List of events related to this Replication Controller.
	EventList common.EventList `json:"eventList"`

	// True when the data contains at least one pod with metrics information, false otherwise.
	HasMetrics bool `json:"hasMetrics"`
}

// ReplicationControllerSpec contains information needed to update replication controller.
type ReplicationControllerSpec struct {
	// Replicas (pods) number in replicas set
	Replicas int32 `json:"replicas"`
}

// GetReplicationControllerDetail returns detailed information about the given replication
// controller in the given namespace.
func GetReplicationControllerDetail(client k8sClient.Interface, heapsterClient client.HeapsterClient,
	namespace, name string) (*ReplicationControllerDetail, error) {
	log.Printf("Getting details of %s replication controller in %s namespace", name, namespace)

	replicationController, err := client.ReplicationControllers(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	podInfo, err := getReplicationControllerPodInfo(client, replicationController, namespace)
	if err != nil {
		return nil, err
	}

	// TODO support pagination
	podList, err := GetReplicationControllerPods(client, heapsterClient, common.NO_PAGINATION,
		name, namespace)
	if err != nil {
		return nil, err
	}

	// TODO support pagination
	eventList, err := GetReplicationControllerEvents(client, common.NO_PAGINATION, namespace, name)
	if err != nil {
		return nil, err
	}

	// TODO support pagination
	serviceList, err := GetReplicationControllerServices(client, common.NO_PAGINATION, namespace,
		name)
	if err != nil {
		return nil, err
	}

	replicationControllerDetail := ToReplicationControllerDetail(replicationController, *podInfo,
		*podList, *eventList, *serviceList)
	return &replicationControllerDetail, nil
}

// TODO(floreks): This should be transactional to make sure that RC will not be deleted without pods
// DeleteReplicationController deletes replication controller with given name in given namespace and
// related pods. Also deletes services related to replication controller if deleteServices is true.
func DeleteReplicationController(client k8sClient.Interface, namespace, name string,
	deleteServices bool) error {

	log.Printf("Deleting %s replication controller from %s namespace", name, namespace)

	if deleteServices {
		if err := DeleteReplicationControllerServices(client, namespace, name); err != nil {
			return err
		}
	}

	pods, err := getRawReplicationControllerPods(client, name, namespace)
	if err != nil {
		return err
	}

	if err := client.ReplicationControllers(namespace).Delete(name); err != nil {
		return err
	}

	for _, pod := range pods {
		if err := client.Pods(namespace).Delete(pod.Name, &api.DeleteOptions{}); err != nil {
			return err
		}
	}

	log.Printf("Successfully deleted %s replication controller from %s namespace", name, namespace)

	return nil
}

// DeleteReplicationControllerServices deletes services related to replication controller with given
// name in given namespace.
func DeleteReplicationControllerServices(client k8sClient.Interface, namespace, name string) error {
	log.Printf("Deleting services related to %s replication controller from %s namespace", name,
		namespace)

	replicationController, err := client.ReplicationControllers(namespace).Get(name)
	if err != nil {
		return err
	}

	labelSelector, err := toLabelSelector(replicationController.Spec.Selector)
	if err != nil {
		return err
	}

	services, err := getServicesForDeletion(client, labelSelector, namespace)
	if err != nil {
		return err
	}

	for _, service := range services {
		if err := client.Services(namespace).Delete(service.Name); err != nil {
			return err
		}
	}

	log.Printf("Successfully deleted services related to %s replication controller from %s namespace",
		name, namespace)

	return nil
}

// UpdateReplicasCount updates number of replicas in Replication Controller based on Replication
// Controller Spec
func UpdateReplicasCount(client k8sClient.Interface, namespace, name string,
	replicationControllerSpec *ReplicationControllerSpec) error {
	log.Printf("Updating replicas count to %d for %s replication controller from %s namespace",
		replicationControllerSpec.Replicas, name, namespace)

	replicationController, err := client.ReplicationControllers(namespace).Get(name)
	if err != nil {
		return err
	}

	replicationController.Spec.Replicas = replicationControllerSpec.Replicas

	_, err = client.ReplicationControllers(namespace).Update(replicationController)
	if err != nil {
		return err
	}

	log.Printf("Successfully updated replicas count to %d for %s replication controller from %s namespace",
		replicationControllerSpec.Replicas, name, namespace)

	return nil
}
