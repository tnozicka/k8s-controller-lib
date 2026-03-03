package naming

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestObjNN(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		obj      metav1.Object
		expected types.NamespacedName
	}{
		{
			name: "namespaced object",
			obj: &metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "default",
			},
			expected: types.NamespacedName{
				Name:      "my-pod",
				Namespace: "default",
			},
		},
		{
			name: "cluster-scoped object",
			obj: &metav1.ObjectMeta{
				Name:      "my-node",
				Namespace: "",
			},
			expected: types.NamespacedName{
				Name:      "my-node",
				Namespace: "",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ObjNN(tc.obj)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestObjKindNN(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		gvk      schema.GroupVersionKind
		obj      metav1.Object
		expected ObjectWithGroupVersionKind
	}{
		{
			name: "non-core group",
			gvk: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			obj: &metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-deploy",
			},
			expected: ObjectWithGroupVersionKind{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "my-deploy",
				},
			},
		},
		{
			name: "core group",
			gvk: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			obj: &metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "my-pod",
			},
			expected: ObjectWithGroupVersionKind{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "kube-system",
					Name:      "my-pod",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ObjKindNN(tc.gvk, tc.obj)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestObjectWithGroupVersionKind_String(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		obj      ObjectWithGroupVersionKind
		expected string
	}{
		{
			name: "non-empty group",
			obj: ObjectWithGroupVersionKind{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "my-deploy",
				},
			},
			expected: "apps.v1.Deployment_default/my-deploy",
		},
		{
			name: "empty group (core)",
			obj: ObjectWithGroupVersionKind{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Pod",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "kube-system",
					Name:      "my-pod",
				},
			},
			expected: "v1.Pod_kube-system/my-pod",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.obj.String()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestObjResourceNN(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		gvr      schema.GroupVersionResource
		obj      metav1.Object
		expected ObjectWithGroupVersionResource
	}{
		{
			name: "non-core group",
			gvr: schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
			obj: &metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-deploy",
			},
			expected: ObjectWithGroupVersionResource{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "my-deploy",
				},
			},
		},
		{
			name: "core group",
			gvr: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			obj: &metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "my-pod",
			},
			expected: ObjectWithGroupVersionResource{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "kube-system",
					Name:      "my-pod",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ObjResourceNN(tc.gvr, tc.obj)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestObjectWithGroupVersionResource_String(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		obj      ObjectWithGroupVersionResource
		expected string
	}{
		{
			name: "non-empty group",
			obj: ObjectWithGroupVersionResource{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "my-deploy",
				},
			},
			expected: "apps.v1.deployments_default/my-deploy",
		},
		{
			name: "empty group (core)",
			obj: ObjectWithGroupVersionResource{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				NamespacedName: types.NamespacedName{
					Namespace: "kube-system",
					Name:      "my-pod",
				},
			},
			expected: "v1.pods_kube-system/my-pod",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.obj.String()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestConciseGVK_String(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		gvk      ConciseGVK
		expected string
	}{
		{
			name: "empty group",
			gvk: ConciseGVK{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			expected: "Pod/v1",
		},
		{
			name: "non-empty group",
			gvk: ConciseGVK{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			expected: "Deployment.apps/v1",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.gvk.String()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestConciseGVR_String(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		gvr      ConciseGVR
		expected string
	}{
		{
			name: "empty group",
			gvr: ConciseGVR{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			expected: "pods/v1",
		},
		{
			name: "non-empty group",
			gvr: ConciseGVR{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
			expected: "deployments.apps/v1",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.gvr.String()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}
