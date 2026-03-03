package controllerhelpers

import (
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_SortedConditions(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name               string
		conditions         []metav1.Condition
		expectedConditions []metav1.Condition
	}{
		{
			name: "conditions are sorted by its known suffix first",
			conditions: []metav1.Condition{
				{Type: "Zoo"},
				{Type: "AlphaProgressing"},
				{Type: "Progressing"},
				{Type: "BetaDegraded"},
				{Type: "Degraded"},
			},
			expectedConditions: []metav1.Condition{
				{Type: "Degraded"},
				{Type: "BetaDegraded"},
				{Type: "Progressing"},
				{Type: "AlphaProgressing"},
				{Type: "Zoo"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Clone(tc.conditions)
			sort.Stable(SortedConditions(got))
			if !apiequality.Semantic.DeepEqual(got, tc.expectedConditions) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedConditions, got))
			}
		})
	}
}

func TestFindStatusConditionStatus(t *testing.T) {
	t.Parallel()

	conditionTrue := metav1.ConditionTrue
	conditionFalse := metav1.ConditionFalse

	tt := []struct {
		name           string
		conditions     []metav1.Condition
		conditionType  string
		expectedStatus *metav1.ConditionStatus
	}{
		{
			name: "returns pointer to status when condition exists with ConditionTrue",
			conditions: []metav1.Condition{
				{
					Type:   "Available",
					Status: metav1.ConditionTrue,
				},
			},
			conditionType:  "Available",
			expectedStatus: &conditionTrue,
		},
		{
			name: "returns pointer to status when condition exists with ConditionFalse",
			conditions: []metav1.Condition{
				{
					Type:   "Degraded",
					Status: metav1.ConditionFalse,
				},
			},
			conditionType:  "Degraded",
			expectedStatus: &conditionFalse,
		},
		{
			name: "returns nil when condition is not found",
			conditions: []metav1.Condition{
				{
					Type:   "Available",
					Status: metav1.ConditionTrue,
				},
			},
			conditionType:  "NonExistent",
			expectedStatus: nil,
		},
		{
			name:           "returns nil when conditions slice is empty",
			conditions:     nil,
			conditionType:  "Available",
			expectedStatus: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FindStatusConditionStatus(tc.conditions, tc.conditionType)
			if !reflect.DeepEqual(got, tc.expectedStatus) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedStatus, got))
			}
		})
	}
}
