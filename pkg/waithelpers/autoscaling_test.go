package waithelpers

import (
	"testing"

	kautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpafake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
)

// TestWaitForVerticalPodAutoscalerState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForVerticalPodAutoscalerState(t *testing.T) {
	t.Parallel()

	vpa := &kautoscalingv1.VerticalPodAutoscaler{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		vpafake.NewSimpleClientset(vpa).AutoscalingV1().VerticalPodAutoscalers(vpa.GetNamespace()),
		vpa.GetName(),
		WaitForVerticalPodAutoscalerState,
	)
}

// TestWaitForVerticalPodAutoscalerDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForVerticalPodAutoscalerDeletion(t *testing.T) {
	t.Parallel()

	vpa := &kautoscalingv1.VerticalPodAutoscaler{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"VerticalPodAutoscaler",
		vpa,
		vpafake.NewSimpleClientset(vpa).AutoscalingV1().VerticalPodAutoscalers(testNamespaceName),
		deleteObject,
		WaitForVerticalPodAutoscalerDeletion,
	)
}

// TestWaitForVerticalPodAutoscalerCheckpointState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForVerticalPodAutoscalerCheckpointState(t *testing.T) {
	t.Parallel()

	vpaCheckpoint := &kautoscalingv1.VerticalPodAutoscalerCheckpoint{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		vpafake.NewSimpleClientset(vpaCheckpoint).AutoscalingV1().VerticalPodAutoscalerCheckpoints(vpaCheckpoint.GetNamespace()),
		vpaCheckpoint.GetName(),
		WaitForVerticalPodAutoscalerCheckpointState,
	)
}

// TestWaitForVerticalPodAutoscalerCheckpointDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForVerticalPodAutoscalerCheckpointDeletion(t *testing.T) {
	t.Parallel()

	vpaCheckpoint := &kautoscalingv1.VerticalPodAutoscalerCheckpoint{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"VerticalPodAutoscalerCheckpoint",
		vpaCheckpoint,
		vpafake.NewSimpleClientset(vpaCheckpoint).AutoscalingV1().VerticalPodAutoscalerCheckpoints(testNamespaceName),
		deleteObject,
		WaitForVerticalPodAutoscalerCheckpointDeletion,
	)
}
