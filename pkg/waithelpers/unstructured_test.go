package waithelpers

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func deleteDynamicObject(ctx context.Context, client dynamic.ResourceInterface, name string) error {
	return client.Delete(ctx, name, metav1.DeleteOptions{})
}

// TestWaitForUnstructuredState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForUnstructuredState(t *testing.T) {
	t.Parallel()

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}

	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]any{
				"name":      "test-cm",
				"namespace": testNamespaceName,
				"annotations": map[string]any{
					testAnnotationKey: "any",
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			gvr: "ConfigMapList",
		},
		obj,
	)

	testWaitForState(
		t,
		dynamicClient.Resource(gvr).Namespace(testNamespaceName),
		obj.GetName(),
		WaitForUnstructuredObjectState,
	)
}

// TestWaitForUnstructuredDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForUnstructuredDeletion(t *testing.T) {
	t.Parallel()

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}

	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]any{
				"name":      "test-cm",
				"namespace": testNamespaceName,
				"uid":       string(testUID),
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			gvr: "ConfigMapList",
		},
		obj,
	)

	testWaitForDeletion(
		t,
		"Unstructured",
		obj,
		dynamicClient.Resource(gvr).Namespace(testNamespaceName),
		deleteDynamicObject,
		WaitForUnstructuredObjectDeletion,
	)
}
