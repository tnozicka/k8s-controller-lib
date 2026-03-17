package controllerhelpers

import (
	"fmt"
	"strings"

	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type SARInterface interface {
	error
	SuggestRBAC() string
}

type SARNotAllowed struct {
	schema.GroupVersionResource
	Verb            string
	ExplicitDenial  bool
	Reason          string
	EvaluationError string
}

var _ SARInterface = &SARNotAllowed{}

func NewSARNotAllowed(obj schema.GroupVersionResource, verb string, denied bool, reason string, evaluationError string) *SARNotAllowed {
	return &SARNotAllowed{
		GroupVersionResource: obj,
		Verb:                 verb,
		ExplicitDenial:       denied,
		Reason:               reason,
		EvaluationError:      evaluationError,
	}
}

func NewSARNotAllowedFromStatus(resource schema.GroupVersionResource, verb string, status authorizationv1.SubjectAccessReviewStatus) *SARNotAllowed {
	return NewSARNotAllowed(
		resource,
		verb,
		status.Denied,
		status.Reason,
		status.EvaluationError,
	)
}

func (s *SARNotAllowed) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("not allowed to %q %q", s.Verb, naming.ConciseGVR(s.GroupVersionResource)))

	if !s.ExplicitDenial {
		sb.WriteString(": authorizer has no opinion on whether to authorize the action")
	}

	sb.WriteString(fmt.Sprintf(": reason %q", s.Reason))

	if len(s.EvaluationError) > 0 {
		sb.WriteString(fmt.Sprintf(" , evaluation error %q", s.EvaluationError))
	}

	return sb.String()
}

func (s *SARNotAllowed) SuggestRBAC() string {
	// TODO: Render a clusterrole and a clusterrolebinding (using role aggregation) for it as a yaml suggestion.
	return fmt.Sprintf(
		"Please adjust the clusterrole for this controller so it can %q resource %q",
		s.Verb,
		naming.ConciseGVR(s.GroupVersionResource),
	)
}
