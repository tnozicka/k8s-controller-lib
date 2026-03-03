package controllerhelpers

import (
	"errors"
	"fmt"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

func RunSync(
	conditions *[]metav1.Condition,
	progressingConditionType string,
	degradedConditionType string,
	observedGeneration int64,
	syncFn func() ([]metav1.Condition, []metav1.Condition, error),
) error {
	var errs []error

	progressingConditions, degradedConditions, syncErr := syncFn()
	if syncErr != nil {
		errs = append(errs, syncErr)
	}

	SetStatusConditionFromError(&degradedConditions, syncErr, degradedConditionType, observedGeneration)
	for _, c := range degradedConditions {
		apimeta.SetStatusCondition(conditions, c)
	}

	mainDegradedCondition, err := AggregateStatusConditions(
		degradedConditions,
		metav1.Condition{
			Type:               degradedConditionType,
			Status:             metav1.ConditionFalse,
			Reason:             naming.StatusConditionAsExpectedReason,
			Message:            "",
			ObservedGeneration: observedGeneration,
		},
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't aggregate degraded conditions for %q: %w", degradedConditionType, err))
	}
	apimeta.SetStatusCondition(conditions, mainDegradedCondition)

	mainProgressingCondition, err := AggregateStatusConditions(
		progressingConditions,
		metav1.Condition{
			Type:               progressingConditionType,
			Status:             metav1.ConditionFalse,
			Reason:             naming.StatusConditionAsExpectedReason,
			Message:            "",
			ObservedGeneration: observedGeneration,
		},
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't aggregate progressing conditions for %q: %w", progressingConditionType, err))
	}
	apimeta.SetStatusCondition(conditions, mainProgressingCondition)

	return errors.Join(errs...)
}
