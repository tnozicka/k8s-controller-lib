package resources

import (
	"maps"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func CmpGroupVersionResources(lhs, rhs schema.GroupVersionResource) int {
	if lhs.Group != rhs.Group {
		return strings.Compare(lhs.Group, rhs.Group)
	}
	if lhs.Version != rhs.Version {
		return strings.Compare(lhs.Version, rhs.Version)
	}
	return strings.Compare(lhs.Resource, rhs.Resource)
}

type ResourceMerger struct {
	preferredVersions map[string]string
}

func NewResourceMerger(preferredVersions map[string]string) *ResourceMerger {
	return &ResourceMerger{
		preferredVersions: preferredVersions,
	}
}

func (vm ResourceMerger) UniqueGroupResources(gvrs ...schema.GroupVersionResource) []schema.GroupVersionResource {
	s := map[schema.GroupResource]schema.GroupVersionResource{}
	for _, gvr := range gvrs {
		existingGVR, exists := s[gvr.GroupResource()]
		if !exists {
			s[gvr.GroupResource()] = gvr
			continue
		}

		// Preferred version only tells us the preferred version for the group,
		// not whether the resource is present in that version.
		preferredGroupVersion, hasPreferredGroup := vm.preferredVersions[gvr.Group]

		// If we have the existing resource in the preferred version, we should keep it.
		if hasPreferredGroup && existingGVR.Version == preferredGroupVersion {
			klog.V(4).InfoS(
				"Pruning duplicate resource version in favour of existing preferred version",
				"DiscardedGVR", gvr,
				"ExistingGVR", existingGVR,
			)
			continue
		}

		// If we have a new resource in the preferred version, it should always win.
		if hasPreferredGroup && gvr.Version == preferredGroupVersion {
			s[gvr.GroupResource()] = gvr
			klog.V(4).InfoS(
				"Pruning duplicate resource version in favour of the new preferred version",
				"NewGVR", gvr,
				"DiscardedGVR", existingGVR,
			)
			continue
		}

		// For the rest, higher semver wins. If not semver, we fall back to simple ordering.
		if semver.Compare(gvr.Version, existingGVR.Version) > 0 {
			s[gvr.GroupResource()] = gvr
			klog.V(4).InfoS(
				"Pruning duplicate resource version in favour of the new version with higher semver",
				"NewGVR", gvr,
				"DiscardedGVR", existingGVR,
			)
		} else {
			klog.V(4).InfoS(
				"Pruning duplicate resource version in favour of existing version",
				"DiscardedGVR", gvr,
				"ExistingGVR", existingGVR,
			)
		}
	}

	res := slices.Collect(maps.Values(s))
	// Ensure stable output, while map iterations are random.
	slices.SortStableFunc(res, CmpGroupVersionResources)

	return res
}
