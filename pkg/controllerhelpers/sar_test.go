package controllerhelpers

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewSARNotAllowed(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		sar         *SARNotAllowed
		expectedSAR *SARNotAllowed
	}{
		{
			name: "creates SARNotAllowed with all fields",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				Verb:            "get",
				ExplicitDenial:  true,
				Reason:          "policy denied",
				EvaluationError: "eval err",
			},
			expectedSAR: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				Verb:            "get",
				ExplicitDenial:  true,
				Reason:          "policy denied",
				EvaluationError: "eval err",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewSARNotAllowed(
				tc.sar.GroupVersionResource,
				tc.sar.Verb,
				tc.sar.ExplicitDenial,
				tc.sar.Reason,
				tc.sar.EvaluationError,
			)
			if !reflect.DeepEqual(got, tc.expectedSAR) {
				t.Errorf("expected and got SARNotAllowed differ:\n%s", cmp.Diff(tc.expectedSAR, got))
			}
		})
	}
}

func TestNewSARNotAllowedFromStatus(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		sar         *SARNotAllowed
		status      authorizationv1.SubjectAccessReviewStatus
		expectedSAR *SARNotAllowed
	}{
		{
			name: "creates SARNotAllowed from SubjectAccessReviewStatus",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				Verb:            "create",
				ExplicitDenial:  false,
				Reason:          "",
				EvaluationError: "",
			},
			status: authorizationv1.SubjectAccessReviewStatus{
				Denied:          true,
				Reason:          "forbidden",
				EvaluationError: "timeout",
			},
			expectedSAR: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				Verb:            "create",
				ExplicitDenial:  true,
				Reason:          "forbidden",
				EvaluationError: "timeout",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewSARNotAllowedFromStatus(tc.sar.GroupVersionResource, tc.sar.Verb, tc.status)
			if !reflect.DeepEqual(got, tc.expectedSAR) {
				t.Errorf("expected and got SARNotAllowed differ:\n%s", cmp.Diff(tc.expectedSAR, got))
			}
		})
	}
}

func TestSARNotAllowed_Error(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name          string
		sar           *SARNotAllowed
		expectedError string
	}{
		{
			name: "explicit denial does not contain no opinion text",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				Verb:            "get",
				ExplicitDenial:  true,
				Reason:          "policy denied",
				EvaluationError: "",
			},
			expectedError: `not allowed to "get" "deployments.apps/v1": reason "policy denied"`,
		},
		{
			name: "not explicit denial contains no opinion text",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				Verb:            "get",
				ExplicitDenial:  false,
				Reason:          "no reason",
				EvaluationError: "",
			},
			expectedError: `not allowed to "get" "deployments.apps/v1": authorizer has no opinion on whether to authorize the action: reason "no reason"`,
		},
		{
			name: "with evaluation error includes it",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				Verb:            "create",
				ExplicitDenial:  true,
				Reason:          "denied",
				EvaluationError: "webhook timeout",
			},
			expectedError: `not allowed to "create" "pods/v1": reason "denied" , evaluation error "webhook timeout"`,
		},
		{
			name: "without evaluation error omits it",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				Verb:            "create",
				ExplicitDenial:  true,
				Reason:          "denied",
				EvaluationError: "",
			},
			expectedError: `not allowed to "create" "pods/v1": reason "denied"`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sar := NewSARNotAllowed(
				tc.sar.GroupVersionResource, tc.sar.Verb, tc.sar.ExplicitDenial,
				tc.sar.Reason, tc.sar.EvaluationError,
			)
			got := sar.Error()
			if got != tc.expectedError {
				t.Errorf("expected and got error differ:\n%s", cmp.Diff(tc.expectedError, got))
			}
		})
	}
}

func TestSARNotAllowed_SuggestRBAC(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name               string
		sar                *SARNotAllowed
		expectedSuggestion string
	}{
		{
			name: "suggests adjusting clusterrole",
			sar: &SARNotAllowed{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				Verb:            "get",
				ExplicitDenial:  true,
				Reason:          "",
				EvaluationError: "",
			},
			expectedSuggestion: `Please adjust the clusterrole for this controller so it can "get" resource "deployments.apps/v1"`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sar := NewSARNotAllowed(
				tc.sar.GroupVersionResource, tc.sar.Verb, tc.sar.ExplicitDenial,
				tc.sar.Reason, tc.sar.EvaluationError,
			)
			got := sar.SuggestRBAC()
			if got != tc.expectedSuggestion {
				t.Errorf("expected and got suggestion differ:\n%s", cmp.Diff(tc.expectedSuggestion, got))
			}
		})
	}
}
