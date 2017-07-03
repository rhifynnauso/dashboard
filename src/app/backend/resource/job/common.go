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

package job

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
)

// The code below allows to perform complex data section on []batch.Job

type JobCell batch.Job

func (self JobCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func (self JobCell) GetResourceSelector() *metricapi.ResourceSelector {
	return &metricapi.ResourceSelector{
		Namespace:    self.ObjectMeta.Namespace,
		ResourceType: api.ResourceKindJob,
		ResourceName: self.ObjectMeta.Name,
		UID:          self.UID,
	}
}

func ToCells(std []batch.Job) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = JobCell(std[i])
	}
	return cells
}

func FromCells(cells []dataselect.DataCell) []batch.Job {
	std := make([]batch.Job, len(cells))
	for i := range std {
		std[i] = batch.Job(cells[i].(JobCell))
	}
	return std
}
