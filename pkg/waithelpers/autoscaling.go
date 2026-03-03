package waithelpers

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	kautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kautoscalingv1client "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
)

func WaitForVerticalPodAutoscalerState(
	ctx context.Context,
	client kautoscalingv1client.VerticalPodAutoscalerInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*kautoscalingv1.VerticalPodAutoscaler) (bool, error),
) (*kautoscalingv1.VerticalPodAutoscaler, error) {
	return WaitForObjectState[*kautoscalingv1.VerticalPodAutoscaler, *kautoscalingv1.VerticalPodAutoscalerList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForVerticalPodAutoscalerDeletion(
	ctx context.Context,
	client kautoscalingv1client.VerticalPodAutoscalerInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*kautoscalingv1.VerticalPodAutoscaler, *kautoscalingv1.VerticalPodAutoscalerList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForVerticalPodAutoscalerCheckpointState(
	ctx context.Context,
	client kautoscalingv1client.VerticalPodAutoscalerCheckpointInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*kautoscalingv1.VerticalPodAutoscalerCheckpoint) (bool, error),
) (*kautoscalingv1.VerticalPodAutoscalerCheckpoint, error) {
	return WaitForObjectState[*kautoscalingv1.VerticalPodAutoscalerCheckpoint, *kautoscalingv1.VerticalPodAutoscalerCheckpointList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForVerticalPodAutoscalerCheckpointDeletion(
	ctx context.Context,
	client kautoscalingv1client.VerticalPodAutoscalerCheckpointInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*kautoscalingv1.VerticalPodAutoscalerCheckpoint, *kautoscalingv1.VerticalPodAutoscalerCheckpointList](
		ctx,
		client,
		nn,
		uid,
	)
}
