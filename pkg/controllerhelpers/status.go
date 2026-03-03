package controllerhelpers

import (
	"errors"
	"fmt"
	"strings"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type DegradedError interface {
	error
	GetReason() string
	GetMessage() string
}

var _ error = DegradedError(nil)

type degradedError struct {
	reason  string
	message string
}

var _ DegradedError = degradedError{}

func NewDegradedError(reason, message string) DegradedError {
	return &degradedError{
		reason:  reason,
		message: message,
	}
}

func (e degradedError) Error() string {
	return fmt.Sprintf("degraded: %s: %v", e.GetReason(), e.GetMessage())
}

func (e degradedError) GetReason() string {
	return e.reason
}

func (e degradedError) GetMessage() string {
	return e.message
}

func SetStatusConditionFromError(
	conditions *[]metav1.Condition,
	err error,
	conditionType string,
	observedGeneration int64,
) {
	if err != nil {
		var degradedErr DegradedError
		if !errors.As(err, &degradedErr) {
			degradedErr = NewDegradedError(
				naming.StatusConditionErrorReason,
				fmt.Sprintf("%v", err),
			)
		}

		apimeta.SetStatusCondition(conditions, metav1.Condition{
			Type:               conditionType,
			Status:             metav1.ConditionTrue,
			Reason:             degradedErr.GetReason(),
			Message:            degradedErr.GetMessage(),
			ObservedGeneration: observedGeneration,
		})
		return
	}

	apimeta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionType,
		Status:             metav1.ConditionFalse,
		Reason:             naming.StatusConditionAsExpectedReason,
		Message:            "",
		ObservedGeneration: observedGeneration,
	})
}

func aggregateStatusConditionInfo(conditions []metav1.Condition) (string, string) {
	reasons := make([]string, 0, len(conditions))
	messages := make([]string, 0, len(conditions))

	for _, c := range conditions {
		reasons = append(reasons, fmt.Sprintf("%s_%s", c.Type, c.Reason))

		for line := range strings.SplitSeq(c.Message, "\n") {
			messages = append(messages, fmt.Sprintf("%s: %s", c.Type, line))
		}
	}

	return strings.Join(reasons, ","), strings.Join(messages, "\n")
}

func AggregateStatusConditions(
	conditions []metav1.Condition,
	condition metav1.Condition,
) (metav1.Condition, error) {
	var defaultVal bool

	switch condition.Status {
	case metav1.ConditionTrue:
		defaultVal = true

	case metav1.ConditionFalse:
		defaultVal = false

	default:
		return metav1.Condition{}, fmt.Errorf("unsupported default value %q for status condition %q", condition.Status, condition.Type)
	}

	var trueConditions, falseConditions, unknownConditions []metav1.Condition
	for _, c := range conditions {
		switch c.Status {
		case metav1.ConditionTrue:
			trueConditions = append(trueConditions, c)

		case metav1.ConditionFalse:
			falseConditions = append(falseConditions, c)

		case metav1.ConditionUnknown:
			unknownConditions = append(unknownConditions, c)

		default:
			return metav1.Condition{}, fmt.Errorf("unsupported condition status %q", c.Status)
		}
	}

	if defaultVal == true && len(falseConditions) > 0 {
		reason, message := aggregateStatusConditionInfo(falseConditions)
		return metav1.Condition{
			Type:               condition.Type,
			Status:             metav1.ConditionFalse,
			Reason:             reason,
			Message:            message,
			ObservedGeneration: condition.ObservedGeneration,
		}, nil
	}

	if defaultVal == false && len(trueConditions) > 0 {
		reason, message := aggregateStatusConditionInfo(trueConditions)
		return metav1.Condition{
			Type:               condition.Type,
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            message,
			ObservedGeneration: condition.ObservedGeneration,
		}, nil
	}

	if len(unknownConditions) > 0 {
		reason, message := aggregateStatusConditionInfo(unknownConditions)
		return metav1.Condition{
			Type:               condition.Type,
			Status:             metav1.ConditionUnknown,
			Reason:             reason,
			Message:            message,
			ObservedGeneration: condition.ObservedGeneration,
		}, nil
	}

	return condition, nil
}

func FindStatusConditionsWithSuffix(conditions []metav1.Condition, suffix string) []metav1.Condition {
	var res []metav1.Condition

	suffixLen := len(suffix)
	for _, c := range conditions {
		if len(c.Type) <= suffixLen {
			continue
		}

		if strings.HasSuffix(c.Type, suffix) {
			res = append(res, c)
		}
	}

	return res
}

func AddAggregatedConditionsUsingSuffix(conditions *[]metav1.Condition, aggregators map[string]metav1.ConditionStatus, generation int64) error {
	for aggregator, defaultStatus := range aggregators {
		cond, err := AggregateStatusConditions(
			FindStatusConditionsWithSuffix(*conditions, aggregator),
			metav1.Condition{
				Type:               aggregator,
				Status:             defaultStatus,
				Reason:             naming.StatusConditionAsExpectedReason,
				Message:            "",
				ObservedGeneration: generation,
			},
		)
		if err != nil {
			return fmt.Errorf("can't aggregate status conditions with suffix %q: %w", aggregator, err)
		}

		apimeta.SetStatusCondition(conditions, cond)
	}

	return nil
}
