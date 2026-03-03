package waithelpers

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestWaitForControllerRevisionState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForControllerRevisionState(t *testing.T) {
	t.Parallel()

	controllerRevision := &appsv1.ControllerRevision{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(controllerRevision).AppsV1().ControllerRevisions(controllerRevision.GetNamespace()),
		controllerRevision.GetName(),
		WaitForControllerRevisionState,
	)
}

// TestWaitForControllerRevisionDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForControllerRevisionDeletion(t *testing.T) {
	t.Parallel()

	controllerRevision := &appsv1.ControllerRevision{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"ControllerRevision",
		controllerRevision,
		fake.NewClientset(controllerRevision).AppsV1().ControllerRevisions(testNamespaceName),
		deleteObject,
		WaitForControllerRevisionDeletion,
	)
}

// TestWaitForDaemonSetState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForDaemonSetState(t *testing.T) {
	t.Parallel()

	daemonSet := &appsv1.DaemonSet{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(daemonSet).AppsV1().DaemonSets(daemonSet.GetNamespace()),
		daemonSet.GetName(),
		WaitForDaemonSetState,
	)
}

// TestWaitForDaemonSetDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForDaemonSetDeletion(t *testing.T) {
	t.Parallel()

	daemonSet := &appsv1.DaemonSet{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"DaemonSet",
		daemonSet,
		fake.NewClientset(daemonSet).AppsV1().DaemonSets(testNamespaceName),
		deleteObject,
		WaitForDaemonSetDeletion,
	)
}

// TestWaitForDeploymentState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForDeploymentState(t *testing.T) {
	t.Parallel()

	deployment := &appsv1.Deployment{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(deployment).AppsV1().Deployments(deployment.GetNamespace()),
		deployment.GetName(),
		WaitForDeploymentState,
	)
}

// TestWaitForDeploymentDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForDeploymentDeletion(t *testing.T) {
	t.Parallel()

	deployment := &appsv1.Deployment{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Deployment",
		deployment,
		fake.NewClientset(deployment).AppsV1().Deployments(testNamespaceName),
		deleteObject,
		WaitForDeploymentDeletion,
	)
}

// TestWaitForReplicaSetState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForReplicaSetState(t *testing.T) {
	t.Parallel()

	replicaSet := &appsv1.ReplicaSet{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(replicaSet).AppsV1().ReplicaSets(replicaSet.GetNamespace()),
		replicaSet.GetName(),
		WaitForReplicaSetState,
	)
}

// TestWaitForReplicaSetDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForReplicaSetDeletion(t *testing.T) {
	t.Parallel()

	replicaSet := &appsv1.ReplicaSet{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"ReplicaSet",
		replicaSet,
		fake.NewClientset(replicaSet).AppsV1().ReplicaSets(testNamespaceName),
		deleteObject,
		WaitForReplicaSetDeletion,
	)
}

// TestWaitForStatefulSetState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForStatefulSetState(t *testing.T) {
	t.Parallel()

	statefulSet := &appsv1.StatefulSet{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(statefulSet).AppsV1().StatefulSets(statefulSet.GetNamespace()),
		statefulSet.GetName(),
		WaitForStatefulSetState,
	)
}

// TestWaitForStatefulSetDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForStatefulSetDeletion(t *testing.T) {
	t.Parallel()

	statefulSet := &appsv1.StatefulSet{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"StatefulSet",
		statefulSet,
		fake.NewClientset(statefulSet).AppsV1().StatefulSets(testNamespaceName),
		deleteObject,
		WaitForStatefulSetDeletion,
	)
}
