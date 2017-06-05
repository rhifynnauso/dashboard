package deployment

import (
	"reflect"
	"testing"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestGetDeploymentEvents(t *testing.T) {
	cases := []struct {
		namespace, name string
		eventList       *v1.EventList
		deployment      *extensions.Deployment
		expectedActions []string
		expected        *common.EventList
	}{
		{
			"ns-1", "dp-1",
			&v1.EventList{Items: []v1.Event{
				{Message: "test-message", ObjectMeta: metaV1.ObjectMeta{
					Name: "ev-1", Namespace: "ns-1", Labels: map[string]string{"app": "test"},
				}},
			}},
			createDeployment("dp-1", "ns-1", "pod-1", map[string]string{"app": "test"},
				map[string]string{"app": "test"}),
			[]string{"list"},
			&common.EventList{
				ListMeta: api.ListMeta{TotalItems: 1},
				Events: []common.Event{{
					TypeMeta: api.TypeMeta{Kind: api.ResourceKindEvent},
					ObjectMeta: api.ObjectMeta{Name: "ev-1", Namespace: "ns-1",
						Labels: map[string]string{"app": "test"}},
					Message: "test-message",
					Type:    v1.EventTypeNormal,
				}}},
		},
	}

	for _, c := range cases {

		fakeClient := fake.NewSimpleClientset(c.eventList, c.deployment)

		actual, _ := GetDeploymentEvents(fakeClient, dataselect.NoDataSelect, c.namespace, c.name)

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetDeploymentEvents(client,%#v, %#v) == \ngot: %#v, \nexpected %#v",
				c.namespace, c.name, actual, c.expected)
		}
	}
}
