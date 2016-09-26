package deployment

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	deploymentutil "k8s.io/kubernetes/pkg/util/deployment"

	"github.com/kubernetes/dashboard/src/app/backend/resource/replicaset"
)

// GetDeploymentEvents returns model events for a deployment with the given name in the given
// namespace
func GetDeploymentOldReplicaSets(client client.Interface, dsQuery *dataselect.DataSelectQuery,
	namespace string, deploymentName string) (*replicaset.ReplicaSetList, error) {

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
		ReplicaSetList: common.GetReplicaSetListChannelWithOptions(client.Extensions(),
			common.NewSameNamespaceQuery(namespace), options, 1),
		PodList: common.GetPodListChannelWithOptions(client,
			common.NewSameNamespaceQuery(namespace), options, 1),
		EventList: common.GetEventListChannelWithOptions(client,
			common.NewSameNamespaceQuery(namespace), options, 1),
	}

	rawRs := <-channels.ReplicaSetList.List
	if err := <-channels.ReplicaSetList.Error; err != nil {
		return nil, err
	}
	rawPods := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}
	rawEvents := <-channels.EventList.List
	if err := <-channels.EventList.Error; err != nil {
		return nil, err
	}

	oldRs, _, err := deploymentutil.FindOldReplicaSets(deployment, rawRs.Items, rawPods)
	if err != nil {
		return nil, err
	}

	oldReplicaSets := make([]extensions.ReplicaSet, len(oldRs))
	for i, replicaSet := range oldRs {
		oldReplicaSets[i] = *replicaSet
	}
	oldReplicaSetList := replicaset.CreateReplicaSetList(oldReplicaSets, rawPods.Items, rawEvents.Items,
		dsQuery, nil)
	return oldReplicaSetList, nil
}
