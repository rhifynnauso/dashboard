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

package thirdpartyresource

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	k8sClient "k8s.io/client-go/kubernetes"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// ThirdPartyResource is a third party resource template.
type ThirdPartyResource struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// ThirdPartyResourceList contains a list of third party resource templates.
type ThirdPartyResourceList struct {
	ListMeta            api.ListMeta         `json:"listMeta"`
	TypeMeta            api.TypeMeta         `json:"typeMeta"`
	ThirdPartyResources []ThirdPartyResource `json:"thirdPartyResources"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetThirdPartyResourceList returns a list of third party resource templates.
func GetThirdPartyResourceList(client k8sClient.Interface, dsQuery *dataselect.DataSelectQuery) (*ThirdPartyResourceList, error) {
	log.Println("Getting list of third party resources")

	channels := &common.ResourceChannels{
		ThirdPartyResourceList: common.GetThirdPartyResourceListChannel(client, 1),
	}

	return GetThirdPartyResourceListFromChannels(channels, dsQuery)
}

// GetThirdPartyResourceListFromChannels returns a list of all third party resources in the cluster
// reading required resource list once from the channels.
func GetThirdPartyResourceListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*ThirdPartyResourceList, error) {
	tprs := <-channels.ThirdPartyResourceList.List
	err := <-channels.ThirdPartyResourceList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	result := getThirdPartyResourceList(tprs.Items, nonCriticalErrors, dsQuery)
	return result, nil
}

func getThirdPartyResourceList(thirdPartyResources []extensions.ThirdPartyResource, nonCriticalErrors []error,
	dsQuery *dataselect.DataSelectQuery) *ThirdPartyResourceList {

	result := &ThirdPartyResourceList{
		ThirdPartyResources: make([]ThirdPartyResource, 0),
		ListMeta:            api.ListMeta{TotalItems: len(thirdPartyResources)},
		Errors:              nonCriticalErrors,
	}

	tprCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(thirdPartyResources), dsQuery)
	thirdPartyResources = fromCells(tprCells)
	result.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, item := range thirdPartyResources {
		result.ThirdPartyResources = append(result.ThirdPartyResources,
			ThirdPartyResource{
				ObjectMeta: api.NewObjectMeta(item.ObjectMeta),
				TypeMeta:   api.NewTypeMeta(api.ResourceKindThirdPartyResource),
			})
	}

	return result
}
