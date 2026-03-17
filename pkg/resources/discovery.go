package resources

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

func FromAPILists(rls []*metav1.APIResourceList, transform func(resource *schema.GroupVersionResource) error) (sets.Set[schema.GroupVersionResource], error) {
	s := sets.New[schema.GroupVersionResource]()
	for _, rl := range rls {
		var gv schema.GroupVersion
		gv, err := schema.ParseGroupVersion(rl.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("can't parse group version %q: %w", rl.GroupVersion, err)
		}

		for i := range rl.APIResources {
			gvr := gv.WithResource(rl.APIResources[i].Name)

			if transform != nil {
				err = transform(&gvr)
				if err != nil {
					return nil, fmt.Errorf("can't transform gvr %q: %w", gvr.String(), err)
				}
			}

			s.Insert(gvr)
		}
	}

	return s, nil
}

func FromAPIList(rl *metav1.APIResourceList, transform func(resource *schema.GroupVersionResource) error) (sets.Set[schema.GroupVersionResource], error) {
	return FromAPILists([]*metav1.APIResourceList{rl}, transform)
}

func FindResourcesForKind(apiResources []*metav1.APIResource, kind string) []*metav1.APIResource {
	var resources []*metav1.APIResource

	for _, r := range apiResources {
		if r.Kind == kind {
			resources = append(resources, r)
		}
	}

	return resources
}

func FindResourceForKind(apiResources []*metav1.APIResource, kind string) (*metav1.APIResource, error) {
	resources := FindResourcesForKind(apiResources, kind)
	if len(resources) == 0 {
		return nil, fmt.Errorf("no resource found for kind %q", kind)
	}

	return resources[0], nil
}
