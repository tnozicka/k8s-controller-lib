package controllerhelpers

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
)

func TestPruneObjects(t *testing.T) {
	t.Parallel()

	cmScheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(cmScheme))

	now := new(metav1.Now())
	nsName := "test-ns"

	newCM := func(name string, deletionTimestamp *metav1.Time) *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         nsName,
				Name:              name,
				UID:               types.UID("uid-" + name),
				DeletionTimestamp: deletionTimestamp,
			},
		}
	}

	const eternalSuffix = "-eternal"

	tt := []struct {
		name               string
		required           []*corev1.ConfigMap
		allObjects         []runtime.Object
		claimedObjects     map[string]*corev1.ConfigMap
		scheme             *runtime.Scheme
		expectedErr        error
		expectedRemaining  []corev1.ConfigMap
		expectedConditions []metav1.Condition
	}{
		{
			name:              "nothing is required and nothing to prune",
			required:          nil,
			allObjects:        nil,
			claimedObjects:    nil,
			scheme:            cmScheme,
			expectedErr:       nil,
			expectedRemaining: nil,
			expectedConditions: []metav1.Condition{
				{
					Type:               "PruningProgressing",
					Status:             metav1.ConditionFalse,
					ObservedGeneration: 1,
					Reason:             "AsExpected",
					Message:            "",
				},
			},
		},
		{
			name: "prunes object that is no longer required",
			required: []*corev1.ConfigMap{
				newCM("keep", nil),
			},
			allObjects: []runtime.Object{
				newCM("keep", nil),
				newCM("prune", nil),
			},
			claimedObjects: map[string]*corev1.ConfigMap{
				"keep":  newCM("keep", nil),
				"prune": newCM("prune", nil),
			},
			scheme:      cmScheme,
			expectedErr: nil,
			expectedRemaining: []corev1.ConfigMap{
				*newCM("keep", nil),
			},
			expectedConditions: []metav1.Condition{
				{
					Type:               "PruningProgressing",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					Reason:             "PruningObject",
					Message:            "Pruning object v1.ConfigMap_test-ns/prune",
				},
			},
		},
		{
			name: "skips delete for objects already being deleted",
			required: []*corev1.ConfigMap{
				newCM("keep", nil),
			},
			allObjects: []runtime.Object{
				newCM("keep", nil),
				newCM("prune", now),
			},
			claimedObjects: map[string]*corev1.ConfigMap{
				"keep":  newCM("keep", nil),
				"prune": newCM("prune", now),
			},
			scheme:      cmScheme,
			expectedErr: nil,
			expectedRemaining: []corev1.ConfigMap{
				*newCM("keep", nil),
				*newCM("prune", now),
			},
			expectedConditions: []metav1.Condition{
				{
					Type:               "PruningProgressing",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					Reason:             "PruningObject",
					Message:            "Pruning object v1.ConfigMap_test-ns/prune",
				},
			},
		},
		{
			name: "prunes object that is no longer required and already deleted",
			required: []*corev1.ConfigMap{
				newCM("keep", nil),
			},
			allObjects: []runtime.Object{
				newCM("keep", nil),
			},
			claimedObjects: map[string]*corev1.ConfigMap{
				"keep":  newCM("keep", nil),
				"prune": newCM("prune", nil),
			},
			scheme:      cmScheme,
			expectedErr: nil,
			expectedRemaining: []corev1.ConfigMap{
				*newCM("keep", nil),
			},
			expectedConditions: []metav1.Condition{
				{
					Type:               "PruningProgressing",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					Reason:             "PruningObject",
					Message:            "Pruning object v1.ConfigMap_test-ns/prune",
				},
			},
		},
		{
			name:     "handles deletion error",
			required: nil,
			allObjects: []runtime.Object{
				newCM("prune", nil),
			},
			claimedObjects: map[string]*corev1.ConfigMap{
				"prune-eternal": newCM("prune-eternal", nil),
			},
			scheme: cmScheme,
			expectedErr: fmt.Errorf(
				`can't prune objects: %w`,
				errors.Join(
					fmt.Errorf(
						`can't delete object v1.ConfigMap_test-ns/prune-eternal: %w`,
						fmt.Errorf("this object is eternal"),
					),
				),
			),
			expectedRemaining: []corev1.ConfigMap{
				*newCM("prune", nil),
			},
			expectedConditions: nil,
		},
		{
			name:     "pruning GVK that's not registered",
			required: nil,
			allObjects: []runtime.Object{
				newCM("prune", nil),
			},
			claimedObjects: map[string]*corev1.ConfigMap{
				"prune": newCM("prune", nil),
			},
			scheme: runtime.NewScheme(),
			expectedErr: fmt.Errorf(
				"can't get GVK for object %T: %w",
				(*corev1.ConfigMap)(nil),
				fmt.Errorf("can't get kind for object %T", (*corev1.ConfigMap)(nil)),
			),
			expectedRemaining: []corev1.ConfigMap{
				*newCM("prune", nil),
			},
			expectedConditions: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewClientset(tc.allObjects...)
			client.PrependReactor("delete", "configmaps", func(action ktesting.Action) (bool, runtime.Object, error) {
				switch action := action.(type) {
				case ktesting.DeleteActionImpl:
					if strings.HasSuffix(action.GetName(), eternalSuffix) {
						return true, nil, errors.New("this object is eternal")
					}
					return false, nil, nil
				default:
					return false, nil, nil
				}
			})

			recorder := record.NewFakeRecorder(10)
			var conditions []metav1.Condition

			err := PruneObjects(
				t.Context(),
				client.CoreV1().ConfigMaps(nsName),
				recorder,
				tc.required,
				tc.claimedObjects,
				1,
				&conditions,
				tc.scheme,
			)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err, cmpopts.EquateErrors()))
			}

			for i := range conditions {
				ltt := conditions[i].LastTransitionTime.Time
				if ltt.IsZero() {
					t.Errorf("expected last transition time to be set on %v", conditions[i].Type)
				}
				conditions[i].LastTransitionTime = metav1.Time{}
			}

			if !apiequality.Semantic.DeepEqual(conditions, tc.expectedConditions) {
				t.Errorf("expected and got conditions differ:\n%s",
					cmp.Diff(tc.expectedConditions, conditions))
			}

			remainingList, listErr := client.CoreV1().ConfigMaps(nsName).List(t.Context(), metav1.ListOptions{})
			if listErr != nil {
				t.Fatalf("can't list remaining ConfigMaps: %v", listErr)
			}
			slices.SortFunc(remainingList.Items, func(a, b corev1.ConfigMap) int {
				return strings.Compare(a.Name, b.Name)
			})
			if !apiequality.Semantic.DeepEqual(remainingList.Items, tc.expectedRemaining) {
				t.Errorf("expected and got remaining differ:\n%s",
					cmp.Diff(tc.expectedRemaining, remainingList.Items))
			}
		})
	}
}
