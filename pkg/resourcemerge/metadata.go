package resourcemerge

import (
	"maps"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isRemovalKey(k string) bool {
	return strings.HasSuffix(k, "-")
}

func isRemovalKey2(k, _ string) bool {
	return isRemovalKey(k)
}

func toRemovalKey(k string) string {
	return k + "-"
}

func mergeMapInPlaceWithoutRemovalKeys(required map[string]string, existing map[string]string) {
	for existingKey, existingValue := range existing {
		// Don't copy removed keys.
		_, isRemoved := required[toRemovalKey(existingKey)]
		if isRemoved {
			continue
		}

		// Copy only keys not present in the required object.
		_, found := required[existingKey]
		if !found {
			required[existingKey] = existingValue
		}
	}

}

// MergeMetadataInPlace merges metadata from existing into the required object.
// Keys already present in the required object are skipped.
// Cleanup keys are removed from the required object.
func MergeMetadataInPlace(required metav1.Object, existing metav1.Object) {
	if required.GetAnnotations() == nil {
		required.SetAnnotations(map[string]string{})
	}
	mergeMapInPlaceWithoutRemovalKeys(required.GetAnnotations(), existing.GetAnnotations())
	maps.DeleteFunc(required.GetAnnotations(), isRemovalKey2)

	if required.GetLabels() == nil {
		required.SetLabels(map[string]string{})
	}
	mergeMapInPlaceWithoutRemovalKeys(required.GetLabels(), existing.GetLabels())
	maps.DeleteFunc(required.GetLabels(), isRemovalKey2)
}
