package controllerhelpers

import (
	"fmt"
	"regexp"
	"sort"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var (
	subTypeRegex                = regexp.MustCompile(`^(?P<name>.*?)(?P<subtype>Progressing|Degraded|Available)?$`)
	subTypeRegexSubtypeKey      = "subtype"
	subTypeRegexSubtypeKeyIndex = func() int {
		i := subTypeRegex.SubexpIndex(subTypeRegexSubtypeKey)
		if i < 0 {
			panic("invalid subtype index")
		}
		return i
	}()
	subTypeRegexNameKey      = "name"
	subTypeRegexNameKeyIndex = func() int {
		i := subTypeRegex.SubexpIndex(subTypeRegexNameKey)
		if i < 0 {
			panic("invalid name index")
		}
		return i
	}()
)

// getConditionTypeSortString returns a string that determines the condition order.
// For conditions with known suffixes (like Progressing, Degraded and Available),
// the suffix is put at the beginning to group the same condition types in a sort.
func getConditionTypeSortString(t string) string {
	var name, subtype string
	lhsMatches := subTypeRegex.FindStringSubmatch(t)
	if lhsMatches == nil {
		klog.ErrorS(fmt.Errorf("invalid condition type %q", t), "")
		return t
	}

	name = lhsMatches[subTypeRegexNameKeyIndex]
	subtype = lhsMatches[subTypeRegexSubtypeKeyIndex]
	return subtype + name
}

type SortedConditions []metav1.Condition

var _ sort.Interface = SortedConditions{}

func (s SortedConditions) Len() int {
	return len(s)
}

func (s SortedConditions) Less(i, j int) bool {
	lhs := s[i]
	lhsSortString := getConditionTypeSortString(lhs.Type)

	rhs := s[j]
	rhsSortString := getConditionTypeSortString(rhs.Type)

	return lhsSortString < rhsSortString
}

func (s SortedConditions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func FindStatusConditionStatus(conditions []metav1.Condition, conditionType string) *metav1.ConditionStatus {
	cond := meta.FindStatusCondition(conditions, conditionType)
	if cond == nil {
		return nil
	}

	return &cond.Status
}
