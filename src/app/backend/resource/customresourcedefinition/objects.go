// Copyright 2017 The Kubernetes Authors.
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

package customresourcedefinition

import (
	"encoding/json"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CustomResourceObject represents a custom resource object.
type CustomResourceObject struct {
	TypeMeta   metav1.TypeMeta   `json:"typeMeta"`
	ObjectMeta metav1.ObjectMeta `json:"objectMeta"`
}

func (r *CustomResourceObject) UnmarshalJSON(data []byte) error {
	tempStruct := &struct {
		metav1.TypeMeta `json:",inline"`
		ObjectMeta      metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}

	err := json.Unmarshal(data, &tempStruct)
	if err != nil {
		return err
	}

	r.TypeMeta = tempStruct.TypeMeta
	r.ObjectMeta = tempStruct.ObjectMeta
	return nil
}

type CustomResourceObjectDetail struct {
	CustomResourceObject `json:",inline"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// CustomResourceObjectList represents crd objects in a namespace.
type CustomResourceObjectList struct {
	TypeMeta metav1.TypeMeta `json:"typeMeta"`
	ListMeta api.ListMeta    `json:"listMeta"`

	// Unordered list of custom resource definitions
	Items []CustomResourceObject `json:"items"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

func (r *CustomResourceObjectList) UnmarshalJSON(data []byte) error {
	tempStruct := &struct {
		metav1.TypeMeta `json:",inline"`
		Items           []CustomResourceObject `json:"items"`
	}{}

	err := json.Unmarshal(data, &tempStruct)
	if err != nil {
		return err
	}

	r.TypeMeta = tempStruct.TypeMeta
	r.Items = tempStruct.Items
	return nil
}

// GetCustomResourceObjectList gets objects for a CR.
func GetCustomResourceObjectList(client apiextensionsclientset.Interface, config *rest.Config, namespace *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery, crdName string) (*CustomResourceObjectList, error) {
	var list *CustomResourceObjectList

	customResourceDefinition, err := client.ApiextensionsV1beta1().
		CustomResourceDefinitions().
		Get(crdName, metav1.GetOptions{})
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	restClient, err := NewRESTClient(config, customResourceDefinition)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	raw, err := restClient.Get().
		Namespace(namespace.ToRequestParam()).
		Resource(customResourceDefinition.Spec.Names.Plural).
		Do().Raw()
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	err = json.Unmarshal(raw, &list)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}
	list.Errors = nonCriticalErrors

	// Return only slice of data, pagination is done here.
	crdObjectCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toObjectCells(list.Items), dsQuery)
	list.Items = fromObjectCells(crdObjectCells)
	list.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for i := range list.Items {
		replaceCRDObjectKind(&list.Items[i], customResourceDefinition.Name)
	}

	return list, nil
}

// GetCustomResourceObjectDetail returns details of a single object in a CR.
func GetCustomResourceObjectDetail(client apiextensionsclientset.Interface, namespace *common.NamespaceQuery, config *rest.Config, crdName string, name string) (*CustomResourceObjectDetail, error) {
	var detail *CustomResourceObjectDetail

	customResourceDefinition, err := client.ApiextensionsV1beta1().
		CustomResourceDefinitions().
		Get(crdName, metav1.GetOptions{})
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	restClient, err := NewRESTClient(config, customResourceDefinition)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	raw, err := restClient.Get().
		Namespace(namespace.ToRequestParam()).
		Resource(customResourceDefinition.Status.AcceptedNames.Plural).
		Name(name).Do().Raw()
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	err = json.Unmarshal(raw, &detail)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}
	detail.Errors = nonCriticalErrors

	replaceCRDObjectKind(&detail.CustomResourceObject, customResourceDefinition.Name)
	return detail, nil
}

// replaceCRDObjectKind sets the object kind to the full name of the CRD.
// E.g. changes "Foo" to "foos.samplecontroller.k8s.io"
func replaceCRDObjectKind(object *CustomResourceObject, kind string) {
	object.TypeMeta.Kind = kind
}
