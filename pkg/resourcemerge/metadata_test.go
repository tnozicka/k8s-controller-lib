package resourcemerge

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeMetadataInPlace(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		required metav1.Object
		existing metav1.Object
		expected metav1.Object
	}{
		{
			name: "existing copied when absent",
			required: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"req-key": "req-val",
				},
				Labels: map[string]string{
					"req-label": "req-label-val",
				},
			},
			existing: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"ext-key": "ext-val",
				},
				Labels: map[string]string{
					"ext-label": "ext-label-val",
				},
			},
			expected: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"req-key": "req-val",
					"ext-key": "ext-val",
				},
				Labels: map[string]string{
					"req-label": "req-label-val",
					"ext-label": "ext-label-val",
				},
			},
		},
		{
			name: "required wins on conflict",
			required: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"key": "required-value",
				},
				Labels: map[string]string{
					"label": "required-label",
				},
			},
			existing: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"key": "existing-value",
				},
				Labels: map[string]string{
					"label": "existing-label",
				},
			},
			expected: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"key": "required-value",
				},
				Labels: map[string]string{
					"label": "required-label",
				},
			},
		},
		{
			name: "removal keys",
			required: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"remove-me-": "",
				},
				Labels: map[string]string{
					"remove-label-": "",
				},
			},
			existing: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"remove-me": "old-val",
					"keep-me":   "keep-val",
				},
				Labels: map[string]string{
					"remove-label": "old-label",
					"keep-label":   "keep-label-val",
				},
			},
			expected: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"keep-me": "keep-val",
				},
				Labels: map[string]string{
					"keep-label": "keep-label-val",
				},
			},
		},
		{
			name: "nil requires",
			required: &metav1.ObjectMeta{
				Annotations: nil,
				Labels:      nil,
			},
			existing: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo": "bar",
				},
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			expected: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo": "bar",
				},
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			name: "mixed operations",
			required: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"shared":     "required-shared",
					"req-only":   "req-only-val",
					"to-remove-": "",
				},
				Labels: map[string]string{
					"shared-label": "required-label",
					"req-label":    "req-label-val",
				},
			},
			existing: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"shared":    "existing-shared",
					"ext-only":  "ext-only-val",
					"to-remove": "should-be-removed",
				},
				Labels: map[string]string{
					"shared-label": "existing-label",
					"ext-label":    "ext-label-val",
				},
			},
			expected: &metav1.ObjectMeta{
				Annotations: map[string]string{
					"shared":   "required-shared",
					"req-only": "req-only-val",
					"ext-only": "ext-only-val",
				},
				Labels: map[string]string{
					"shared-label": "required-label",
					"req-label":    "req-label-val",
					"ext-label":    "ext-label-val",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.required
			MergeMetadataInPlace(got, tc.existing)

			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(got, tc.required))
			}
		})
	}
}
