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

package node

import (
	"reflect"
	"testing"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/metric"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	api "k8s.io/client-go/pkg/api/v1"
)

func TestGetNodeList(t *testing.T) {
	cases := []struct {
		node     *api.Node
		expected *NodeList
	}{
		{
			&api.Node{
				ObjectMeta: metaV1.ObjectMeta{Name: "test-node"},
				Spec: api.NodeSpec{
					Unschedulable: true,
				},
			},
			&NodeList{
				ListMeta: common.ListMeta{
					TotalItems: 1,
				},
				CumulativeMetrics: make([]metric.Metric, 0),
				Nodes: []Node{{
					ObjectMeta: common.ObjectMeta{Name: "test-node"},
					TypeMeta:   common.TypeMeta{Kind: common.ResourceKindNode},
					Ready:      "Unknown",
					AllocatedResources: NodeAllocatedResources{
						CPURequests:            0,
						CPURequestsFraction:    0,
						CPULimits:              0,
						CPULimitsFraction:      0,
						CPUCapacity:            0,
						MemoryRequests:         0,
						MemoryRequestsFraction: 0,
						MemoryLimits:           0,
						MemoryLimitsFraction:   0,
						MemoryCapacity:         0,
						AllocatedPods:          0,
						PodCapacity:            0,
					},
				},
				},
			},
		},
	}

	for _, c := range cases {
		fakeClient := fake.NewSimpleClientset(c.node)
		fakeHeapsterClient := FakeHeapsterClient{client: *fake.NewSimpleClientset()}
		actual, _ := GetNodeList(fakeClient, dataselect.NoDataSelect, fakeHeapsterClient)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetNodeList() == \ngot: %#v, \nexpected %#v", actual, c.expected)
		}
	}
}
