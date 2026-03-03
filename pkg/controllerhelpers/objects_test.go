package controllerhelpers

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
)

func TestConvertNamesIntoLocalObjectReferences(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		names          []string
		expectedResult []corev1.LocalObjectReference
	}{
		{
			name:           "zero names returns empty slice",
			names:          []string{},
			expectedResult: []corev1.LocalObjectReference{},
		},
		{
			name:  "one name",
			names: []string{"alpha"},
			expectedResult: []corev1.LocalObjectReference{
				{
					Name: "alpha",
				},
			},
		},
		{
			name:  "multiple names",
			names: []string{"alpha", "beta", "gamma"},
			expectedResult: []corev1.LocalObjectReference{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
				{
					Name: "gamma",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ConvertNamesIntoLocalObjectReferences(tc.names...)
			if !apiequality.Semantic.DeepEqual(got, tc.expectedResult) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedResult, got))
			}
		})
	}
}

func TestMintManagedDataToSlice(t *testing.T) {
	t.Parallel()

	const namespaceName = "test-ns"
	const wrongNamespaceName = "test-ns" + "-wrong"

	owner := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "owner",
			Namespace: "test-ns",
			UID:       types.UID("owner-uid"),
		},
	}

	ownerGVK := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	}

	tt := []struct {
		name         string
		objs         []kubetypes.Object
		expectedObjs []kubetypes.Object
		expectedErr  error
	}{
		{
			name:        "nil slice",
			objs:        nil,
			expectedErr: nil,
		},
		{
			name: "multiple objects",
			objs: []kubetypes.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   namespaceName,
						Name:        "foo",
						Annotations: nil,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespaceName,
						Name:      "bar",
						Annotations: map[string]string{
							"existing-key": "existing-value",
						},
					},
				},
			},
			expectedObjs: []kubetypes.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   namespaceName,
						Name:        "foo",
						Annotations: map[string]string{},
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespaceName,
						Name:      "bar",
						Annotations: map[string]string{
							"existing-key": "existing-value",
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "namespace mismatch produces error",
			objs: []kubetypes.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   wrongNamespaceName,
						Name:        "bar",
						Annotations: nil,
					},
				},
			},
			expectedObjs: nil,
			expectedErr: fmt.Errorf(
				`can't mint managed data for object "test-ns-wrong/bar": %w`,
				errors.New(`object namespace "test-ns-wrong" differs from minted namespace "test-ns"`),
			),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := MintManagedDataToSlice(
				tc.objs,
				"test-ns",
				labels.Set{"app": "test"},
				owner,
				ownerGVK,
			)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}
		})
	}
}
