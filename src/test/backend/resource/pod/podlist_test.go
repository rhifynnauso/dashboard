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
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func TestGetRestartCount(t *testing.T) {
	cases := []struct {
		pod      api.Pod
		expected int
	}{
		{
			api.Pod{}, 0,
		},
		{
			api.Pod{
				Status: api.PodStatus{
					ContainerStatuses: []api.ContainerStatus{
						{
							Name:         "container-1",
							RestartCount: 1,
						},
					},
				},
			},
			1,
		},
		{
			api.Pod{
				Status: api.PodStatus{
					ContainerStatuses: []api.ContainerStatus{
						{
							Name:         "container-1",
							RestartCount: 3,
						},
						{
							Name:         "container-2",
							RestartCount: 2,
						},
					},
				},
			},
			5,
		},
	}
	for _, c := range cases {
		actual := getRestartCount(c.pod)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("AppendEvents(%#v) == %#v, expected %#v", c.pod, actual, c.expected)
		}
	}
}
