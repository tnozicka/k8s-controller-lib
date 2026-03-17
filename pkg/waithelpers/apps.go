package waithelpers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func WaitForControllerRevisionState(
	ctx context.Context,
	client appsv1client.ControllerRevisionInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*appsv1.ControllerRevision) (bool, error),
) (*appsv1.ControllerRevision, error) {
	return WaitForObjectState[*appsv1.ControllerRevision, *appsv1.ControllerRevisionList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForControllerRevisionDeletion(
	ctx context.Context,
	client appsv1client.ControllerRevisionInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*appsv1.ControllerRevision, *appsv1.ControllerRevisionList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForDaemonSetState(
	ctx context.Context,
	client appsv1client.DaemonSetInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*appsv1.DaemonSet) (bool, error),
) (*appsv1.DaemonSet, error) {
	return WaitForObjectState[*appsv1.DaemonSet, *appsv1.DaemonSetList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForDaemonSetDeletion(
	ctx context.Context,
	client appsv1client.DaemonSetInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*appsv1.DaemonSet, *appsv1.DaemonSetList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForDeploymentState(
	ctx context.Context,
	client appsv1client.DeploymentInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*appsv1.Deployment) (bool, error),
) (*appsv1.Deployment, error) {
	return WaitForObjectState[*appsv1.Deployment, *appsv1.DeploymentList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForDeploymentDeletion(
	ctx context.Context,
	client appsv1client.DeploymentInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*appsv1.Deployment, *appsv1.DeploymentList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForReplicaSetState(
	ctx context.Context,
	client appsv1client.ReplicaSetInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*appsv1.ReplicaSet) (bool, error),
) (*appsv1.ReplicaSet, error) {
	return WaitForObjectState[*appsv1.ReplicaSet, *appsv1.ReplicaSetList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForReplicaSetDeletion(
	ctx context.Context,
	client appsv1client.ReplicaSetInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*appsv1.ReplicaSet, *appsv1.ReplicaSetList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForStatefulSetState(
	ctx context.Context,
	client appsv1client.StatefulSetInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*appsv1.StatefulSet) (bool, error),
) (*appsv1.StatefulSet, error) {
	return WaitForObjectState[*appsv1.StatefulSet, *appsv1.StatefulSetList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForStatefulSetDeletion(
	ctx context.Context,
	client appsv1client.StatefulSetInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*appsv1.StatefulSet, *appsv1.StatefulSetList](
		ctx,
		client,
		nn,
		uid,
	)
}
