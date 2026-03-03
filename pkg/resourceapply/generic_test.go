package resourceapply

import (
	"errors"
	"fmt"
	"iter"
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/fake"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
)

// Picking a generic object like *corev1.Service that already has a generated
// clientset and a lister makes the test easier to read,
// but we shouldn't get bound to a particular type in the tests.
type objectType = corev1.Service

func newControllerRef() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "fake.group/v1",
		Kind:       "FakeObject",
		Name:       "Controller",
		UID:        "controller-uid",
		Controller: new(true),
	}
}

func newService(clusterIP string) *objectType {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       metav1.NamespaceDefault,
			Name:            "test",
			OwnerReferences: nil,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: clusterIP,
		},
	}
}

func newObjectWithoutControllerRef(v string) *objectType {
	return newService(v)
}

func newObjectWithControllerRef(v string) *objectType {
	obj := newObjectWithoutControllerRef(v)
	obj.SetOwnerReferences([]metav1.OwnerReference{
		newControllerRef(),
	})
	return obj
}

func newHashAnnotator() HashAnnotator {
	return NewHashAnnotator("test/managed-hash")
}

func withHash[T kubetypes.Object](obj T) T {
	utilruntime.Must(newHashAnnotator().SetHashAnnotationWithCleanup(obj))
	return obj
}

func getHash(obj kubetypes.Object) string {
	return newHashAnnotator().GetHashAnnotation(withHash(obj))
}

func withExplicitHash[T kubetypes.Object](obj T, hash string) T {
	obj = withHash(obj)
	obj.GetAnnotations()[newHashAnnotator().GetHashAnnotationKey()] = hash
	return obj
}

func sanitizeTestObject(obj *objectType) {
	if obj == nil {
		return
	}
	obj.APIVersion = ""
	obj.Kind = ""
	obj.SetManagedFields(nil)
}

func TestApplyGeneric(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	tt := []struct {
		name                      string
		existing                  []runtime.Object
		cache                     []runtime.Object // nil cache means it should be autofilled from the client.
		required                  *objectType
		forceOwnership            bool
		allowMissingControllerRef bool
		expectedObject            *objectType
		expectedChanged           bool
		expectedErr               error
		expectedEvents            []string
	}{
		{
			name:                      "creates a new object when none exists",
			existing:                  nil,
			cache:                     nil,
			required:                  newObjectWithControllerRef("foo"),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            withHash(newObjectWithControllerRef("foo")),
			expectedChanged:           true,
			expectedErr:               nil,
			expectedEvents:            []string{"Normal V1ServiceCreated v1.Service_default/test created"},
		},
		{
			name: "does nothing if the same object already exists",
			existing: []runtime.Object{
				withHash(newObjectWithControllerRef("foo")),
			},
			cache:                     nil,
			required:                  newObjectWithControllerRef("foo"),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            withHash(newObjectWithControllerRef("foo")),
			expectedChanged:           false,
			expectedErr:               nil,
			expectedEvents:            nil,
		},
		{
			name: "updates the object if the existing one doesn't have the hash",
			existing: []runtime.Object{
				newObjectWithControllerRef("foo"),
			},
			cache:                     nil,
			required:                  newObjectWithControllerRef("bar"),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            withHash(newObjectWithControllerRef("bar")),
			expectedChanged:           true,
			expectedErr:               nil,
			expectedEvents:            []string{"Normal V1ServiceUpdated v1.Service_default/test updated"},
		},
		{
			name: "updates the object if the existing one has a different hash",
			existing: []runtime.Object{
				withHash(newObjectWithControllerRef("foo")),
			},
			cache:                     nil,
			required:                  newObjectWithControllerRef("bar"),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            withHash(newObjectWithControllerRef("bar")),
			expectedChanged:           true,
			expectedErr:               nil,
			expectedEvents:            []string{"Normal V1ServiceUpdated v1.Service_default/test updated"},
		},
		{
			name: "won't update object if the existing one has the same hash",
			existing: []runtime.Object{
				withExplicitHash(
					newObjectWithControllerRef("bar"),
					getHash(newObjectWithControllerRef("foo")),
				),
			},
			cache:                     nil,
			required:                  withHash(newObjectWithControllerRef("foo")),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject: withExplicitHash(
				newObjectWithControllerRef("bar"),
				getHash(newObjectWithControllerRef("foo")),
			),
			expectedChanged: false,
			expectedErr:     nil,
			expectedEvents:  nil,
		},
		{
			name: "specifying no RV will use the one from the existing object",
			existing: []runtime.Object{
				func() runtime.Object {
					obj := newObjectWithControllerRef("foo")
					obj.SetResourceVersion("42")
					return withHash(obj)
				}(),
			},
			cache: nil,
			required: func() *objectType {
				obj := newObjectWithControllerRef("bar")
				obj.SetResourceVersion("")
				return withHash(obj)
			}(),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject: func() *objectType {
				obj := newObjectWithControllerRef("bar")
				obj.SetResourceVersion("42")
				return withHash(obj)
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal V1ServiceUpdated v1.Service_default/test updated"},
		},
		{
			name:     "update fails if the object is missing but we still see it in the cache",
			existing: nil,
			cache: []runtime.Object{
				withHash(newObjectWithControllerRef("foo")),
			},
			required:                  withHash(newObjectWithControllerRef("bar")),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            nil,
			expectedChanged:           false,
			expectedErr:               fmt.Errorf(`can't update object "v1.Service_default/test": %w`, apierrors.NewNotFound(corev1.Resource("services"), "test")),
			expectedEvents:            []string{`Warning UpdateV1ServiceFailed Failed to update v1.Service_default/test: services "test" not found`},
		},
		{
			name:                      "forbids to create the object without a controllerRef",
			existing:                  nil,
			cache:                     nil,
			required:                  newObjectWithoutControllerRef("foo"),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            nil,
			expectedChanged:           false,
			expectedErr:               fmt.Errorf(`"v1.Service_default/test" is missing controllerRef`),
			expectedEvents:            nil,
		},
		{
			name:                      "allows creating the object without a controllerRef if allowMissingControllerRef option is true",
			existing:                  nil,
			cache:                     nil,
			required:                  withHash(newObjectWithoutControllerRef("foo")),
			forceOwnership:            false,
			allowMissingControllerRef: true,
			expectedObject:            withHash(newObjectWithoutControllerRef("foo")),
			expectedChanged:           true,
			expectedErr:               nil,
			expectedEvents:            []string{"Normal V1ServiceCreated v1.Service_default/test created"},
		},
		{
			name: "update fails if the existing object has no ownerRef",
			existing: []runtime.Object{
				withHash(newObjectWithoutControllerRef("foo")),
			},
			cache:                     nil,
			required:                  withHash(newObjectWithControllerRef("bar")),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            nil,
			expectedChanged:           false,
			expectedErr:               fmt.Errorf(`"v1.Service_default/test" isn't controlled by us`),
			expectedEvents:            []string{`Warning UpdateV1ServiceFailed Failed to update v1.Service_default/test: "v1.Service_default/test" isn't controlled by us`},
		},
		{
			name: "update succeeds if the existing object has no ownerRef but has forceOwnership option",
			existing: []runtime.Object{
				withHash(newObjectWithoutControllerRef("foo")),
			},
			cache:                     nil,
			required:                  withHash(newObjectWithControllerRef("bar")),
			forceOwnership:            true,
			allowMissingControllerRef: false,
			expectedObject:            withHash(newObjectWithControllerRef("bar")),
			expectedChanged:           true,
			expectedErr:               nil,
			expectedEvents:            []string{"Normal V1ServiceUpdated v1.Service_default/test updated"},
		},
		{
			name: "update fails if the existing object is owned by someone else",
			existing: []runtime.Object{
				func() runtime.Object {
					obj := withHash(newObjectWithControllerRef("foo"))
					metav1.GetControllerOfNoCopy(obj).UID += "-changed"
					return obj
				}(),
			},
			cache:                     nil,
			required:                  withHash(newObjectWithControllerRef("foo")),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject:            nil,
			expectedChanged:           false,
			expectedErr:               fmt.Errorf(`"v1.Service_default/test" isn't controlled by us`),
			expectedEvents:            []string{`Warning UpdateV1ServiceFailed Failed to update v1.Service_default/test: "v1.Service_default/test" isn't controlled by us`},
		},
		{
			name: "forced update fails if the existing object is owned by someone else",
			existing: []runtime.Object{
				func() runtime.Object {
					obj := withHash(newObjectWithControllerRef("foo"))
					metav1.GetControllerOfNoCopy(obj).UID += "-changed"
					return obj
				}(),
			},
			cache:                     nil,
			required:                  withHash(newObjectWithControllerRef("foo")),
			forceOwnership:            true,
			allowMissingControllerRef: false,
			expectedObject:            nil,
			expectedChanged:           false,
			expectedErr:               fmt.Errorf(`"v1.Service_default/test" isn't controlled by us`),
			expectedEvents:            []string{`Warning UpdateV1ServiceFailed Failed to update v1.Service_default/test: "v1.Service_default/test" isn't controlled by us`},
		},
		{
			name: "all labels and annotations are kept when the hash matches",
			existing: []runtime.Object{
				func() runtime.Object {
					var obj kubetypes.Object
					obj = newObjectWithControllerRef("foo")
					obj.SetAnnotations(map[string]string{
						"da-1":  "da-alpha",
						"da-2":  "da-beta",
						"da-3-": "",
					})
					obj.SetLabels(map[string]string{
						"dl-1":  "dl-alpha",
						"dl-2":  "dl-beta",
						"dl-3-": "",
					})
					obj = withHash(obj)
					obj.GetAnnotations()["da-1"] = "da-alpha-changed"
					obj.GetAnnotations()["da-3"] = "da-resurrected"
					obj.GetAnnotations()["a-custom"] = "a-custom-value"
					obj.GetLabels()["dl-1"] = "dl-alpha-changed"
					obj.GetLabels()["dl-3"] = "dl-resurrected"
					obj.GetLabels()["l-custom"] = "l-custom-value"
					return obj
				}(),
			},
			cache: nil,
			required: func() *objectType {
				obj := newObjectWithControllerRef("foo")
				obj.SetAnnotations(map[string]string{
					"da-1":  "da-alpha",
					"da-2":  "da-beta",
					"da-3-": "",
				})
				obj.SetLabels(map[string]string{
					"dl-1":  "dl-alpha",
					"dl-2":  "dl-beta",
					"dl-3-": "",
				})
				return obj
			}(),
			forceOwnership:            false,
			allowMissingControllerRef: false,
			expectedObject: func() *objectType {
				obj := newObjectWithControllerRef("foo")
				obj.SetAnnotations(map[string]string{
					"da-1":  "da-alpha",
					"da-2":  "da-beta",
					"da-3-": "",
				})
				obj.SetLabels(map[string]string{
					"dl-1":  "dl-alpha",
					"dl-2":  "dl-beta",
					"dl-3-": "",
				})
				obj = withHash(obj)
				obj.GetAnnotations()["da-1"] = "da-alpha-changed"
				obj.GetAnnotations()["da-3"] = "da-resurrected"
				obj.GetAnnotations()["a-custom"] = "a-custom-value"
				obj.GetLabels()["dl-1"] = "dl-alpha-changed"
				obj.GetLabels()["dl-3"] = "dl-resurrected"
				obj.GetLabels()["l-custom"] = "l-custom-value"
				return obj
			}(),
			expectedChanged: false,
			expectedErr:     nil,
			expectedEvents:  nil,
		},
		{
			name: "only managed label and annotation keys are updated when the hash changes",
			existing: []runtime.Object{
				func() runtime.Object {
					var obj kubetypes.Object
					obj = newObjectWithControllerRef("foo")
					obj.SetAnnotations(map[string]string{
						"da-1":  "da-alpha",
						"da-2":  "da-beta",
						"da-3-": "",
					})
					obj.SetLabels(map[string]string{
						"dl-1":  "dl-alpha",
						"dl-2":  "dl-beta",
						"dl-3-": "",
					})
					obj = withHash(obj)
					obj.GetAnnotations()["da-1"] = "da-alpha-changed"
					obj.GetAnnotations()["a-custom"] = "a-custom-value"
					obj.GetLabels()["dl-1"] = "dl-alpha-changed"
					obj.GetLabels()["l-custom"] = "l-custom-value"
					return obj
				}(),
			},
			cache: nil,
			required: func() *objectType {
				obj := newObjectWithControllerRef("foo")
				obj.SetAnnotations(map[string]string{
					"da-1":  "da-alpha-new",
					"da-2":  "da-beta-new",
					"da-3-": "",
				})
				obj.SetLabels(map[string]string{
					"dl-1":  "dl-alpha-new",
					"dl-2":  "dl-beta-new",
					"dl-3-": "",
				})
				obj = withHash(obj)
				return obj
			}(),
			forceOwnership:            true,
			allowMissingControllerRef: false,
			expectedObject: func() *objectType {
				obj := newObjectWithControllerRef("foo")
				obj.SetAnnotations(map[string]string{
					"da-1":  "da-alpha-new",
					"da-2":  "da-beta-new",
					"da-3-": "",
				})
				obj.SetLabels(map[string]string{
					"dl-1":  "dl-alpha-new",
					"dl-2":  "dl-beta-new",
					"dl-3-": "",
				})
				obj = withHash(obj)
				delete(obj.GetAnnotations(), "da-3-")
				obj.GetAnnotations()["a-custom"] = "a-custom-value"
				delete(obj.GetLabels(), "dl-3-")
				obj.GetLabels()["l-custom"] = "l-custom-value"
				return obj
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal V1ServiceUpdated v1.Service_default/test updated"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewClientset(tc.existing...)

			// We need to test that ApplyObject is reentrant and gives identical results.
			iterations := 2
			if tc.expectedErr != nil {
				iterations = 1
			}
			for iteration := range iterations {
				t.Run("", func(t *testing.T) {
					ctx := t.Context()

					recorder := record.NewFakeRecorder(10)

					indexer := cache.NewIndexer(
						cache.MetaNamespaceKeyFunc,
						cache.Indexers{
							cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
						},
					)
					svcLister := corev1listers.NewServiceLister(indexer)

					if tc.cache != nil {
						for _, obj := range tc.cache {
							err := indexer.Add(obj)
							if err != nil {
								t.Fatal(err)
							}
						}
					} else {
						objList, err := client.CoreV1().Services(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{
							LabelSelector: labels.Everything().String(),
						})
						if err != nil {
							t.Fatal(err)
						}

						for i := range objList.Items {
							err = indexer.Add(&objList.Items[i])
							if err != nil {
								t.Fatal(err)
							}
						}
					}

					got, gotChanged, gotErr := ApplyGeneric[*corev1.Service](
						ctx,
						&Applier{
							scheme:        scheme,
							hashAnnotator: newHashAnnotator(),
							recorder:      recorder,
						},
						ApplyControl[*corev1.Service]{
							GetCached: svcLister.Services(metav1.NamespaceDefault).Get,
							Create:    client.CoreV1().Services(metav1.NamespaceDefault).Create,
							Update:    client.CoreV1().Services(metav1.NamespaceDefault).Update,
						},
						tc.required,
						ApplyOptions{
							ForceOwnership:            tc.forceOwnership,
							AllowMissingControllerRef: tc.allowMissingControllerRef,
							FieldValidation:           ptr.To(metav1.FieldValidationStrict),
						},
					)
					if !reflect.DeepEqual(gotErr, tc.expectedErr) {
						t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, gotErr, cmpopts.EquateErrors()))
					}

					sanitizeTestObject(got)
					if !equality.Semantic.DeepEqual(got, tc.expectedObject) {
						t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedObject, got))
					}

					if got != nil {
						createdObj, err := client.CoreV1().Services(got.Namespace).Get(ctx, got.Name, metav1.GetOptions{})
						if err != nil {
							t.Error(err)
						}
						sanitizeTestObject(createdObj)
						if !equality.Semantic.DeepEqual(createdObj, got) {
							t.Errorf("created and returned objects differ:\n%s", cmp.Diff(createdObj, got))
						}
					}

					if iteration == 0 {
						if gotChanged != tc.expectedChanged {
							t.Errorf("expected %t, got %t", tc.expectedChanged, gotChanged)
						}
					} else {
						if gotChanged {
							t.Errorf("object changed in iteration %d", iteration)
						}
					}

					close(recorder.Events)
					var gotEvents []string
					for e := range recorder.Events {
						gotEvents = append(gotEvents, e)
					}
					if iteration == 0 {
						if !reflect.DeepEqual(gotEvents, tc.expectedEvents) {
							t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedEvents, gotEvents))
						}
					} else {
						if len(gotEvents) > 0 {
							t.Errorf("unexpected events: %v", gotEvents)
						}
					}
				})
			}
		})
	}
}

func chanSeq[T any](ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	}
}

func TestApplyOptions_GetFieldValidation(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		options  ApplyOptions
		expected string
	}{
		{
			name:     "nil FieldValidation returns Strict",
			options:  ApplyOptions{},
			expected: metav1.FieldValidationStrict,
		},
		{
			name:     "set FieldValidation returns value",
			options:  ApplyOptions{FieldValidation: new(metav1.FieldValidationWarn)},
			expected: metav1.FieldValidationWarn,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.options.GetFieldValidation()
			if got != tc.expected {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestApplier_reportDeleteEvent(t *testing.T) {
	tt := []struct {
		name           string
		obj            kubetypes.Object
		gvk            schema.GroupVersionKind
		operationErr   error
		expectedEvents []string
	}{
		{
			name:           "success",
			obj:            newService("172.16.0.1"),
			gvk:            corev1.SchemeGroupVersion.WithKind("Service"),
			operationErr:   nil,
			expectedEvents: []string{"Normal V1ServiceDeleted v1.Service_default/test deleted"},
		},
		{
			name:           "error",
			obj:            newService("172.16.0.1"),
			gvk:            corev1.SchemeGroupVersion.WithKind("Service"),
			operationErr:   errors.New("can't delete service"),
			expectedEvents: []string{"Warning DeleteV1ServiceFailed Failed to delete v1.Service_default/test: can't delete service"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := record.NewFakeRecorder(10)
			a := &Applier{
				recorder: recorder,
			}
			a.reportDeleteEvent(tc.obj, tc.gvk, tc.operationErr)
			close(recorder.Events)
			events := slices.Collect(chanSeq(recorder.Events))
			if !equality.Semantic.DeepEqual(events, tc.expectedEvents) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedEvents, events))
			}
		})
	}
}
