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

package deployment

import (
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// GetDeploymentEvents returns model events for a deployment with the given name in the given
// namespace
func GetDeploymentEvents(client client.Interface, dsQuery *dataselect.DataSelectQuery, namespace string, deploymentName string) (
	*common.EventList, error) {

	deployment, err := client.Extensions().Deployments(namespace).Get(deploymentName)
	if err != nil {
		return nil, err
	}

	selector, err := unversioned.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}

	options := api.ListOptions{LabelSelector: selector}
	channels := &common.ResourceChannels{
		EventList: common.GetEventListChannelWithOptions(client,
			common.NewSameNamespaceQuery(namespace), options, 1),
	}

	eventRaw := <-channels.EventList.List
	if err := <-channels.EventList.Error; err != nil {
		return nil, err
	}
	dpEvents := eventRaw.Items
	if !event.IsTypeFilled(dpEvents) {
		dpEvents = event.FillEventsType(dpEvents)
	}

	eventList := event.CreateEventList(dpEvents, dsQuery)

	return &eventList, nil
}
