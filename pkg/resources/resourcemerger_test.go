package resources

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCmpGroupVersionResources(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		lhs      schema.GroupVersionResource
		rhs      schema.GroupVersionResource
		expected int
	}{
		{
			name:     "equal GVRs",
			lhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			rhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			expected: 0,
		},
		{
			name:     "lhs Group less than rhs Group",
			lhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			rhs:      schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "deployments"},
			expected: -1,
		},
		{
			name:     "lhs Group greater than rhs Group",
			lhs:      schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "deployments"},
			rhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			expected: 1,
		},
		{
			name:     "same Group, lhs Version less than rhs Version",
			lhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			rhs:      schema.GroupVersionResource{Group: "apps", Version: "v2", Resource: "deployments"},
			expected: -1,
		},
		{
			name:     "same Group, lhs Version greater than rhs Version",
			lhs:      schema.GroupVersionResource{Group: "apps", Version: "v2", Resource: "deployments"},
			rhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			expected: 1,
		},
		{
			name:     "same Group and Version, lhs Resource less than rhs Resource",
			lhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"},
			rhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			expected: -1,
		},
		{
			name:     "same Group and Version, lhs Resource greater than rhs Resource",
			lhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
			rhs:      schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			expected: 1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CmpGroupVersionResources(tc.lhs, tc.rhs)
			if got != tc.expected {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestNewResourceMerger(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name              string
		preferredVersions map[string]string
		expected          *ResourceMerger
	}{
		{
			name: "with preferred versions",
			preferredVersions: map[string]string{
				"apps": "v1",
			},
			expected: &ResourceMerger{
				preferredVersions: map[string]string{
					"apps": "v1",
				},
			},
		},
		{
			name:              "nil preferred versions",
			preferredVersions: nil,
			expected: &ResourceMerger{
				preferredVersions: nil,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewResourceMerger(tc.preferredVersions)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestResourceMerger_UniqueGroupResources(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name              string
		preferredVersions map[string]string
		gvrs              []schema.GroupVersionResource
		expectedGVRs      []schema.GroupVersionResource
	}{
		{
			name:              "no duplicates returns all",
			preferredVersions: map[string]string{},
			gvrs: []schema.GroupVersionResource{
				{Group: "apps", Version: "v1", Resource: "deployments"},
				{Group: "", Version: "v1", Resource: "pods"},
			},
			expectedGVRs: []schema.GroupVersionResource{
				{Group: "", Version: "v1", Resource: "pods"},
				{Group: "apps", Version: "v1", Resource: "deployments"},
			},
		},
		{
			name: "preferred version wins when existing is preferred",
			preferredVersions: map[string]string{
				"apps": "v1",
			},
			gvrs: []schema.GroupVersionResource{
				{Group: "apps", Version: "v1", Resource: "deployments"},
				{Group: "apps", Version: "v1beta1", Resource: "deployments"},
			},
			expectedGVRs: []schema.GroupVersionResource{
				{Group: "apps", Version: "v1", Resource: "deployments"},
			},
		},
		{
			name: "preferred version wins when new is preferred",
			preferredVersions: map[string]string{
				"apps": "v1",
			},
			gvrs: []schema.GroupVersionResource{
				{Group: "apps", Version: "v1beta1", Resource: "deployments"},
				{Group: "apps", Version: "v1", Resource: "deployments"},
			},
			expectedGVRs: []schema.GroupVersionResource{
				{Group: "apps", Version: "v1", Resource: "deployments"},
			},
		},
		{
			name:              "higher semver wins when no preferred version",
			preferredVersions: map[string]string{},
			gvrs: []schema.GroupVersionResource{
				{Group: "example.com", Version: "v1.0.0", Resource: "widgets"},
				{Group: "example.com", Version: "v2.0.0", Resource: "widgets"},
			},
			expectedGVRs: []schema.GroupVersionResource{
				{Group: "example.com", Version: "v2.0.0", Resource: "widgets"},
			},
		},
		{
			name:              "non-semver fallback keeps existing",
			preferredVersions: map[string]string{},
			gvrs: []schema.GroupVersionResource{
				{Group: "example.com", Version: "alpha", Resource: "widgets"},
				{Group: "example.com", Version: "beta", Resource: "widgets"},
			},
			expectedGVRs: []schema.GroupVersionResource{
				{Group: "example.com", Version: "alpha", Resource: "widgets"},
			},
		},
		{
			name:              "empty input returns empty",
			preferredVersions: map[string]string{},
			gvrs:              []schema.GroupVersionResource{},
			expectedGVRs:      nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rm := NewResourceMerger(tc.preferredVersions)
			got := rm.UniqueGroupResources(tc.gvrs...)

			if !reflect.DeepEqual(got, tc.expectedGVRs) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedGVRs, got))
			}
		})
	}
}
