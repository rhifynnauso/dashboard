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

package namespace

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// NamespaceList contains a list of namespaces in the cluster.
type NamespaceList struct {
	ListMeta common.ListMeta `json:"listMeta"`

	// Unordered list of Namespaces.
	Namespaces []Namespace `json:"namespaces"`
}

// Namespace is a presentation layer view of Kubernetes namespaces. This means it is namespace plus
// additional augumented data we can get from other sources.
type Namespace struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	// Phase is the current lifecycle phase of the namespace.
	Phase api.NamespacePhase `json:"phase"`
}

// GetNamespaceList returns a list of all namespaces in the cluster.
func GetNamespaceList(client *client.Client, pQuery *common.PaginationQuery) (*NamespaceList,
	error) {
	log.Printf("Getting namespace list")

	namespaces, err := client.Namespaces().List(api.ListOptions{
		LabelSelector: labels.Everything(),
		FieldSelector: fields.Everything(),
	})

	if err != nil {
		return nil, err
	}

	return toNamespaceList(namespaces.Items, pQuery), nil
}

func toNamespaceList(namespaces []api.Namespace, pQuery *common.PaginationQuery) *NamespaceList {
	namespaceList := &NamespaceList{
		Namespaces: make([]Namespace, 0),
		ListMeta:   common.ListMeta{TotalItems: len(namespaces)},
	}

	namespaces = paginate(namespaces, pQuery)

	for _, namespace := range namespaces {
		namespaceList.Namespaces = append(namespaceList.Namespaces, toNamespace(namespace))
	}

	return namespaceList
}

func toNamespace(namespace api.Namespace) Namespace {
	return Namespace{
		ObjectMeta: common.NewObjectMeta(namespace.ObjectMeta),
		TypeMeta:   common.NewTypeMeta(common.ResourceKindNamespace),
		Phase:      namespace.Status.Phase,
	}
}
