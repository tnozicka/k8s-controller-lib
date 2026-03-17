package waithelpers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

func WaitForComponentStatusState(
	ctx context.Context,
	client corev1client.ComponentStatusInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.ComponentStatus) (bool, error),
) (*corev1.ComponentStatus, error) {
	return WaitForObjectState[*corev1.ComponentStatus, *corev1.ComponentStatusList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForComponentStatusDeletion(
	ctx context.Context,
	client corev1client.ComponentStatusInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.ComponentStatus, *corev1.ComponentStatusList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForConfigMapState(
	ctx context.Context,
	client corev1client.ConfigMapInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.ConfigMap) (bool, error),
) (*corev1.ConfigMap, error) {
	return WaitForObjectState[*corev1.ConfigMap, *corev1.ConfigMapList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForConfigMapDeletion(
	ctx context.Context,
	client corev1client.ConfigMapInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.ConfigMap, *corev1.ConfigMapList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForEndpointsState(
	ctx context.Context,
	client corev1client.EndpointsInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Endpoints) (bool, error),
) (*corev1.Endpoints, error) {
	return WaitForObjectState[*corev1.Endpoints, *corev1.EndpointsList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForEndpointsDeletion(
	ctx context.Context,
	client corev1client.EndpointsInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Endpoints, *corev1.EndpointsList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForEventState(
	ctx context.Context,
	client corev1client.EventInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Event) (bool, error),
) (*corev1.Event, error) {
	return WaitForObjectState[*corev1.Event, *corev1.EventList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForEventDeletion(
	ctx context.Context,
	client corev1client.EventInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Event, *corev1.EventList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForLimitRangeState(
	ctx context.Context,
	client corev1client.LimitRangeInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.LimitRange) (bool, error),
) (*corev1.LimitRange, error) {
	return WaitForObjectState[*corev1.LimitRange, *corev1.LimitRangeList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForLimitRangeDeletion(
	ctx context.Context,
	client corev1client.LimitRangeInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.LimitRange, *corev1.LimitRangeList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForNamespaceState(
	ctx context.Context,
	client corev1client.NamespaceInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Namespace) (bool, error),
) (*corev1.Namespace, error) {
	return WaitForObjectState[*corev1.Namespace, *corev1.NamespaceList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForNamespaceDeletion(
	ctx context.Context,
	client corev1client.NamespaceInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Namespace, *corev1.NamespaceList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForNodeState(
	ctx context.Context,
	client corev1client.NodeInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Node) (bool, error),
) (*corev1.Node, error) {
	return WaitForObjectState[*corev1.Node, *corev1.NodeList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForNodeDeletion(
	ctx context.Context,
	client corev1client.NodeInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Node, *corev1.NodeList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForPersistentVolumeState(
	ctx context.Context,
	client corev1client.PersistentVolumeInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.PersistentVolume) (bool, error),
) (*corev1.PersistentVolume, error) {
	return WaitForObjectState[*corev1.PersistentVolume, *corev1.PersistentVolumeList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForPersistentVolumeDeletion(
	ctx context.Context,
	client corev1client.PersistentVolumeInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.PersistentVolume, *corev1.PersistentVolumeList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForPersistentVolumeClaimState(
	ctx context.Context,
	client corev1client.PersistentVolumeClaimInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.PersistentVolumeClaim) (bool, error),
) (*corev1.PersistentVolumeClaim, error) {
	return WaitForObjectState[*corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaimList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForPersistentVolumeClaimDeletion(
	ctx context.Context,
	client corev1client.PersistentVolumeClaimInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaimList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForPodState(
	ctx context.Context,
	client corev1client.PodInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Pod) (bool, error),
) (*corev1.Pod, error) {
	return WaitForObjectState[*corev1.Pod, *corev1.PodList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForPodDeletion(
	ctx context.Context,
	client corev1client.PodInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Pod, *corev1.PodList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForPodTemplateState(
	ctx context.Context,
	client corev1client.PodTemplateInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.PodTemplate) (bool, error),
) (*corev1.PodTemplate, error) {
	return WaitForObjectState[*corev1.PodTemplate, *corev1.PodTemplateList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForPodTemplateDeletion(
	ctx context.Context,
	client corev1client.PodTemplateInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.PodTemplate, *corev1.PodTemplateList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForReplicationControllerState(
	ctx context.Context,
	client corev1client.ReplicationControllerInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.ReplicationController) (bool, error),
) (*corev1.ReplicationController, error) {
	return WaitForObjectState[*corev1.ReplicationController, *corev1.ReplicationControllerList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForReplicationControllerDeletion(
	ctx context.Context,
	client corev1client.ReplicationControllerInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.ReplicationController, *corev1.ReplicationControllerList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForResourceQuotaState(
	ctx context.Context,
	client corev1client.ResourceQuotaInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.ResourceQuota) (bool, error),
) (*corev1.ResourceQuota, error) {
	return WaitForObjectState[*corev1.ResourceQuota, *corev1.ResourceQuotaList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForResourceQuotaDeletion(
	ctx context.Context,
	client corev1client.ResourceQuotaInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.ResourceQuota, *corev1.ResourceQuotaList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForSecretState(
	ctx context.Context,
	client corev1client.SecretInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Secret) (bool, error),
) (*corev1.Secret, error) {
	return WaitForObjectState[*corev1.Secret, *corev1.SecretList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForSecretDeletion(
	ctx context.Context,
	client corev1client.SecretInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Secret, *corev1.SecretList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForServiceState(
	ctx context.Context,
	client corev1client.ServiceInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.Service) (bool, error),
) (*corev1.Service, error) {
	return WaitForObjectState[*corev1.Service, *corev1.ServiceList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForServiceDeletion(
	ctx context.Context,
	client corev1client.ServiceInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.Service, *corev1.ServiceList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForServiceAccountState(
	ctx context.Context,
	client corev1client.ServiceAccountInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*corev1.ServiceAccount) (bool, error),
) (*corev1.ServiceAccount, error) {
	return WaitForObjectState[*corev1.ServiceAccount, *corev1.ServiceAccountList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForServiceAccountDeletion(
	ctx context.Context,
	client corev1client.ServiceAccountInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*corev1.ServiceAccount, *corev1.ServiceAccountList](
		ctx,
		client,
		nn,
		uid,
	)
}
