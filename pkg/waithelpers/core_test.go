package waithelpers

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestWaitForComponentStatusState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForComponentStatusState(t *testing.T) {
	t.Parallel()

	componentStatus := &corev1.ComponentStatus{ObjectMeta: newTestMetaClusterScoped("any")}
	testWaitForState(
		t,
		fake.NewClientset(componentStatus).CoreV1().ComponentStatuses(),
		componentStatus.GetName(),
		WaitForComponentStatusState,
	)
}

// TestWaitForComponentStatusDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForComponentStatusDeletion(t *testing.T) {
	t.Parallel()

	componentStatus := &corev1.ComponentStatus{ObjectMeta: newTestMetaClusterScoped("")}
	testWaitForDeletion(
		t,
		"ComponentStatus",
		componentStatus,
		fake.NewClientset(componentStatus).CoreV1().ComponentStatuses(),
		deleteObject,
		WaitForComponentStatusDeletion,
	)
}

// TestWaitForConfigMapState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForConfigMapState(t *testing.T) {
	t.Parallel()

	cm := &corev1.ConfigMap{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(cm).CoreV1().ConfigMaps(cm.GetNamespace()),
		cm.GetName(),
		WaitForConfigMapState,
	)
}

// TestWaitForConfigMapDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForConfigMapDeletion(t *testing.T) {
	t.Parallel()

	cm := &corev1.ConfigMap{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"ConfigMap",
		cm,
		fake.NewClientset(cm).CoreV1().ConfigMaps(testNamespaceName),
		deleteObject,
		WaitForConfigMapDeletion,
	)
}

// TestWaitForEndpointsState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForEndpointsState(t *testing.T) {
	t.Parallel()

	endpoints := &corev1.Endpoints{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(endpoints).CoreV1().Endpoints(endpoints.GetNamespace()),
		endpoints.GetName(),
		WaitForEndpointsState,
	)
}

// TestWaitForEndpointsDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForEndpointsDeletion(t *testing.T) {
	t.Parallel()

	endpoints := &corev1.Endpoints{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Endpoints",
		endpoints,
		fake.NewClientset(endpoints).CoreV1().Endpoints(testNamespaceName),
		deleteObject,
		WaitForEndpointsDeletion,
	)
}

// TestWaitForEventState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForEventState(t *testing.T) {
	t.Parallel()

	event := &corev1.Event{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(event).CoreV1().Events(event.GetNamespace()),
		event.GetName(),
		WaitForEventState,
	)
}

// TestWaitForEventDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForEventDeletion(t *testing.T) {
	t.Parallel()

	event := &corev1.Event{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Event",
		event,
		fake.NewClientset(event).CoreV1().Events(testNamespaceName),
		deleteObject,
		WaitForEventDeletion,
	)
}

// TestWaitForLimitRangeState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForLimitRangeState(t *testing.T) {
	t.Parallel()

	limitRange := &corev1.LimitRange{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(limitRange).CoreV1().LimitRanges(limitRange.GetNamespace()),
		limitRange.GetName(),
		WaitForLimitRangeState,
	)
}

// TestWaitForLimitRangeDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForLimitRangeDeletion(t *testing.T) {
	t.Parallel()

	limitRange := &corev1.LimitRange{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"LimitRange",
		limitRange,
		fake.NewClientset(limitRange).CoreV1().LimitRanges(testNamespaceName),
		deleteObject,
		WaitForLimitRangeDeletion,
	)
}

// TestWaitForNamespaceState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForNamespaceState(t *testing.T) {
	t.Parallel()

	namespace := &corev1.Namespace{ObjectMeta: newTestMetaClusterScoped("any")}
	testWaitForState(
		t,
		fake.NewClientset(namespace).CoreV1().Namespaces(),
		namespace.GetName(),
		WaitForNamespaceState,
	)
}

// TestWaitForNamespaceDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForNamespaceDeletion(t *testing.T) {
	t.Parallel()

	namespace := &corev1.Namespace{ObjectMeta: newTestMetaClusterScoped("")}
	testWaitForDeletion(
		t,
		"Namespace",
		namespace,
		fake.NewClientset(namespace).CoreV1().Namespaces(),
		deleteObject,
		WaitForNamespaceDeletion,
	)
}

// TestWaitForNodeState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForNodeState(t *testing.T) {
	t.Parallel()

	node := &corev1.Node{ObjectMeta: newTestMetaClusterScoped("any")}
	testWaitForState(
		t,
		fake.NewClientset(node).CoreV1().Nodes(),
		node.GetName(),
		WaitForNodeState,
	)
}

// TestWaitForNodeDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForNodeDeletion(t *testing.T) {
	t.Parallel()

	node := &corev1.Node{ObjectMeta: newTestMetaClusterScoped("")}
	testWaitForDeletion(
		t,
		"Node",
		node,
		fake.NewClientset(node).CoreV1().Nodes(),
		deleteObject,
		WaitForNodeDeletion,
	)
}

// TestWaitForPersistentVolumeState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForPersistentVolumeState(t *testing.T) {
	t.Parallel()

	pv := &corev1.PersistentVolume{ObjectMeta: newTestMetaClusterScoped("any")}
	testWaitForState(
		t,
		fake.NewClientset(pv).CoreV1().PersistentVolumes(),
		pv.GetName(),
		WaitForPersistentVolumeState,
	)
}

// TestWaitForPersistentVolumeDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForPersistentVolumeDeletion(t *testing.T) {
	t.Parallel()

	pv := &corev1.PersistentVolume{ObjectMeta: newTestMetaClusterScoped("")}
	testWaitForDeletion(
		t,
		"PersistentVolume",
		pv,
		fake.NewClientset(pv).CoreV1().PersistentVolumes(),
		deleteObject,
		WaitForPersistentVolumeDeletion,
	)
}

// TestWaitForPersistentVolumeClaimState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForPersistentVolumeClaimState(t *testing.T) {
	t.Parallel()

	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(pvc).CoreV1().PersistentVolumeClaims(pvc.GetNamespace()),
		pvc.GetName(),
		WaitForPersistentVolumeClaimState,
	)
}

// TestWaitForPersistentVolumeClaimDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForPersistentVolumeClaimDeletion(t *testing.T) {
	t.Parallel()

	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"PersistentVolumeClaim",
		pvc,
		fake.NewClientset(pvc).CoreV1().PersistentVolumeClaims(testNamespaceName),
		deleteObject,
		WaitForPersistentVolumeClaimDeletion,
	)
}

// TestWaitForPodState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForPodState(t *testing.T) {
	t.Parallel()

	pod := &corev1.Pod{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(pod).CoreV1().Pods(pod.GetNamespace()),
		pod.GetName(),
		WaitForPodState,
	)
}

// TestWaitForPodDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForPodDeletion(t *testing.T) {
	t.Parallel()

	pod := &corev1.Pod{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Pod",
		pod,
		fake.NewClientset(pod).CoreV1().Pods(testNamespaceName),
		deleteObject,
		WaitForPodDeletion,
	)
}

// TestWaitForPodTemplateState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForPodTemplateState(t *testing.T) {
	t.Parallel()

	podTemplate := &corev1.PodTemplate{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(podTemplate).CoreV1().PodTemplates(podTemplate.GetNamespace()),
		podTemplate.GetName(),
		WaitForPodTemplateState,
	)
}

// TestWaitForPodTemplateDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForPodTemplateDeletion(t *testing.T) {
	t.Parallel()

	podTemplate := &corev1.PodTemplate{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"PodTemplate",
		podTemplate,
		fake.NewClientset(podTemplate).CoreV1().PodTemplates(testNamespaceName),
		deleteObject,
		WaitForPodTemplateDeletion,
	)
}

// TestWaitForReplicationControllerState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForReplicationControllerState(t *testing.T) {
	t.Parallel()

	rc := &corev1.ReplicationController{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(rc).CoreV1().ReplicationControllers(rc.GetNamespace()),
		rc.GetName(),
		WaitForReplicationControllerState,
	)
}

// TestWaitForReplicationControllerDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForReplicationControllerDeletion(t *testing.T) {
	t.Parallel()

	rc := &corev1.ReplicationController{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"ReplicationController",
		rc,
		fake.NewClientset(rc).CoreV1().ReplicationControllers(testNamespaceName),
		deleteObject,
		WaitForReplicationControllerDeletion,
	)
}

// TestWaitForResourceQuotaState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForResourceQuotaState(t *testing.T) {
	t.Parallel()

	rq := &corev1.ResourceQuota{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(rq).CoreV1().ResourceQuotas(rq.GetNamespace()),
		rq.GetName(),
		WaitForResourceQuotaState,
	)
}

// TestWaitForResourceQuotaDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForResourceQuotaDeletion(t *testing.T) {
	t.Parallel()

	rq := &corev1.ResourceQuota{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"ResourceQuota",
		rq,
		fake.NewClientset(rq).CoreV1().ResourceQuotas(testNamespaceName),
		deleteObject,
		WaitForResourceQuotaDeletion,
	)
}

// TestWaitForSecretState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForSecretState(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(secret).CoreV1().Secrets(secret.GetNamespace()),
		secret.GetName(),
		WaitForSecretState,
	)
}

// TestWaitForSecretDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForSecretDeletion(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Secret",
		secret,
		fake.NewClientset(secret).CoreV1().Secrets(testNamespaceName),
		deleteObject,
		WaitForSecretDeletion,
	)
}

// TestWaitForServiceState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForServiceState(t *testing.T) {
	t.Parallel()

	service := &corev1.Service{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(service).CoreV1().Services(service.GetNamespace()),
		service.GetName(),
		WaitForServiceState,
	)
}

// TestWaitForServiceDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForServiceDeletion(t *testing.T) {
	t.Parallel()

	service := &corev1.Service{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Service",
		service,
		fake.NewClientset(service).CoreV1().Services(testNamespaceName),
		deleteObject,
		WaitForServiceDeletion,
	)
}

// TestWaitForServiceAccountState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForServiceAccountState(t *testing.T) {
	t.Parallel()

	sa := &corev1.ServiceAccount{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(sa).CoreV1().ServiceAccounts(sa.GetNamespace()),
		sa.GetName(),
		WaitForServiceAccountState,
	)
}

// TestWaitForServiceAccountDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForServiceAccountDeletion(t *testing.T) {
	t.Parallel()

	sa := &corev1.ServiceAccount{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"ServiceAccount",
		sa,
		fake.NewClientset(sa).CoreV1().ServiceAccounts(testNamespaceName),
		deleteObject,
		WaitForServiceAccountDeletion,
	)
}
