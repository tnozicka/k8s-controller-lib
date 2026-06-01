package kubernetes

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/fake"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func TestApplyDeployment(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(appsv1.AddToScheme(scheme))

	annotator := resourceapply.NewHashAnnotator("internal.k8s-controller-lib/managed-hash")

	sanitize := func(obj *appsv1.Deployment) {
		if obj == nil {
			return
		}
		obj.APIVersion = ""
		obj.Kind = ""
		obj.SetManagedFields(nil)
	}

	newDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       "default",
				Name:            "test",
				ResourceVersion: "42",
				Labels:          map[string]string{},
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller:         new(true),
						UID:                "abcdefgh",
						APIVersion:         "example.com/v1",
						Kind:               "MyKind",
						Name:               "owner",
						BlockOwnerDeletion: new(true),
					},
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: new(int32(3)),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "c1",
								Image: "repo/image:latest",
							},
						},
					},
				},
			},
		}
	}

	newDeploymentWithHash := func() *appsv1.Deployment {
		d := newDeployment()
		utilruntime.Must(annotator.SetHashAnnotationWithCleanup(d))
		return d
	}

	tt := []struct {
		name               string
		existing           []runtime.Object
		cache              []runtime.Object // nil cache means autofill from the client
		required           *appsv1.Deployment
		forceOwnership     bool
		expectedDeployment *appsv1.Deployment
		expectedChanged    bool
		expectedErr        error
		expectedEvents     []string
	}{
		{
			name:               "creates a new Deployment when there is none",
			existing:           nil,
			required:           newDeployment(),
			expectedDeployment: newDeploymentWithHash(),
			expectedChanged:    true,
			expectedErr:        nil,
			expectedEvents:     []string{"Normal AppsV1DeploymentCreated apps.v1.Deployment_default/test created"},
		},
		{
			name: "does nothing if the same Deployment already exists",
			existing: []runtime.Object{
				newDeploymentWithHash(),
			},
			required:           newDeployment(),
			expectedDeployment: newDeploymentWithHash(),
			expectedChanged:    false,
			expectedErr:        nil,
			expectedEvents:     nil,
		},
		{
			name: "does nothing if the same Deployment already exists and required one has the hash",
			existing: []runtime.Object{
				newDeploymentWithHash(),
			},
			required:           newDeploymentWithHash(),
			expectedDeployment: newDeploymentWithHash(),
			expectedChanged:    false,
			expectedErr:        nil,
			expectedEvents:     nil,
		},
		{
			name: "updates the Deployment if it exists without the hash",
			existing: []runtime.Object{
				newDeployment(),
			},
			required:           newDeployment(),
			expectedDeployment: newDeploymentWithHash(),
			expectedChanged:    true,
			expectedErr:        nil,
			expectedEvents:     []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name:     "fails to create the Deployment without a controllerRef",
			existing: nil,
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.OwnerReferences = nil
				return deployment
			}(),
			expectedDeployment: nil,
			expectedChanged:    false,
			expectedErr:        fmt.Errorf(`"apps.v1.Deployment_default/test" is missing controllerRef`),
			expectedEvents:     nil,
		},
		{
			name: "updates the Deployment if replicas differ",
			existing: []runtime.Object{
				newDeployment(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Replicas = new(*deployment.Spec.Replicas + 1)
				return deployment
			}(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Replicas = new(*deployment.Spec.Replicas + 1)
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name: "updates the Deployment if labels differ",
			existing: []runtime.Object{
				newDeploymentWithHash(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Labels["foo"] = "bar"
				return deployment
			}(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Labels["foo"] = "bar"
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name: "updates the Deployment if an image differs",
			existing: []runtime.Object{
				newDeploymentWithHash(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name: "won't update the Deployment if an admission changes the Deployment",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeploymentWithHash()
					// Simulate admission by changing a value after the hash is computed.
					deployment.Spec.Template.Spec.Containers[0].Image += "-admissionchange"
					return deployment
				}(),
			},
			required: newDeployment(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeploymentWithHash()
				// Simulate admission by changing a value after the hash is computed.
				deployment.Spec.Template.Spec.Containers[0].Image += "-admissionchange"
				return deployment
			}(),
			expectedChanged: false,
			expectedErr:     nil,
			expectedEvents:  nil,
		},
		{
			// We test propagating the RV from required in all the other tests.
			name: "specifying no RV will use the one from the existing object",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeploymentWithHash()
					deployment.ResourceVersion = "21"
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.ResourceVersion = ""
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.ResourceVersion = "21"
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name:     "update fails if the Deployment is missing but we still see it in the cache",
			existing: nil,
			cache: []runtime.Object{
				newDeploymentWithHash(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			expectedDeployment: nil,
			expectedChanged:    false,
			expectedErr:        fmt.Errorf(`can't update object %q: %w`, "apps.v1.Deployment_default/test", apierrors.NewNotFound(appsv1.Resource("deployments"), "test")),
			expectedEvents:     []string{`Warning UpdateAppsV1DeploymentFailed Failed to update apps.v1.Deployment_default/test: deployments.apps "test" not found`},
		},
		{
			name: "update fails if the existing object has no ownerRef",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.OwnerReferences = nil
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			expectedDeployment: nil,
			expectedChanged:    false,
			expectedErr:        fmt.Errorf(`"apps.v1.Deployment_default/test" isn't controlled by us`),
			expectedEvents:     []string{`Warning UpdateAppsV1DeploymentFailed Failed to update apps.v1.Deployment_default/test: "apps.v1.Deployment_default/test" isn't controlled by us`},
		},
		{
			name: "forced update succeeds if the existing object has no ownerRef",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.OwnerReferences = nil
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			forceOwnership: true,
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name: "update succeeds to replace ownerRef kind",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.OwnerReferences[0].Kind = "WrongKind"
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				return deployment
			}(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name: "update fails if the existing object is owned by someone else",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.OwnerReferences[0].UID = "42"
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			expectedDeployment: nil,
			expectedChanged:    false,
			expectedErr:        fmt.Errorf(`"apps.v1.Deployment_default/test" isn't controlled by us`),
			expectedEvents:     []string{`Warning UpdateAppsV1DeploymentFailed Failed to update apps.v1.Deployment_default/test: "apps.v1.Deployment_default/test" isn't controlled by us`},
		},
		{
			name: "forced update fails if the existing object is owned by someone else",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.OwnerReferences[0].UID = "42"
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Spec.Containers[0].Image += "-rc.0"
				return deployment
			}(),
			forceOwnership:     true,
			expectedDeployment: nil,
			expectedChanged:    false,
			expectedErr:        fmt.Errorf(`"apps.v1.Deployment_default/test" isn't controlled by us`),
			expectedEvents:     []string{`Warning UpdateAppsV1DeploymentFailed Failed to update apps.v1.Deployment_default/test: "apps.v1.Deployment_default/test" isn't controlled by us`},
		},
		{
			name: "all label and annotation keys are kept when the hash matches",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.Annotations = map[string]string{
						"a-1":  "a-alpha",
						"a-2":  "a-beta",
						"a-3-": "",
					}
					deployment.Labels = map[string]string{
						"l-1":  "l-alpha",
						"l-2":  "l-beta",
						"l-3-": "",
					}
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					deployment.Annotations["a-1"] = "a-alpha-changed"
					deployment.Annotations["a-3"] = "a-resurrected"
					deployment.Annotations["a-custom"] = "custom-value"
					deployment.Labels["l-1"] = "l-alpha-changed"
					deployment.Labels["l-3"] = "l-resurrected"
					deployment.Labels["l-custom"] = "custom-value"
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Annotations = map[string]string{
					"a-1":  "a-alpha",
					"a-2":  "a-beta",
					"a-3-": "",
				}
				deployment.Labels = map[string]string{
					"l-1":  "l-alpha",
					"l-2":  "l-beta",
					"l-3-": "",
				}
				return deployment
			}(),
			forceOwnership: false,
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Annotations = map[string]string{
					"a-1":  "a-alpha",
					"a-2":  "a-beta",
					"a-3-": "",
				}
				deployment.Labels = map[string]string{
					"l-1":  "l-alpha",
					"l-2":  "l-beta",
					"l-3-": "",
				}
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				deployment.Annotations["a-1"] = "a-alpha-changed"
				deployment.Annotations["a-3"] = "a-resurrected"
				deployment.Annotations["a-custom"] = "custom-value"
				deployment.Labels["l-1"] = "l-alpha-changed"
				deployment.Labels["l-3"] = "l-resurrected"
				deployment.Labels["l-custom"] = "custom-value"
				return deployment
			}(),
			expectedChanged: false,
			expectedErr:     nil,
			expectedEvents:  nil,
		},
		{
			name: "only managed label and annotation keys are updated when the hash changes",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.Annotations = map[string]string{
						"a-1":  "a-alpha",
						"a-2":  "a-beta",
						"a-3-": "a-resurrected",
					}
					deployment.Labels = map[string]string{
						"l-1":  "l-alpha",
						"l-2":  "l-beta",
						"l-3-": "l-resurrected",
					}
					utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
					deployment.Annotations["a-1"] = "a-alpha-changed"
					deployment.Annotations["a-custom"] = "a-custom-value"
					deployment.Labels["l-1"] = "l-alpha-changed"
					deployment.Labels["l-custom"] = "l-custom-value"
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Annotations = map[string]string{
					"a-1":  "a-alpha-x",
					"a-2":  "a-beta-x",
					"a-3-": "",
				}
				deployment.Labels = map[string]string{
					"l-1":  "l-alpha-x",
					"l-2":  "l-beta-x",
					"l-3-": "",
				}
				return deployment
			}(),
			forceOwnership: true,
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Annotations = map[string]string{
					"a-1":  "a-alpha-x",
					"a-2":  "a-beta-x",
					"a-3-": "",
				}
				deployment.Labels = map[string]string{
					"l-1":  "l-alpha-x",
					"l-2":  "l-beta-x",
					"l-3-": "",
				}
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				delete(deployment.Annotations, "a-3-")
				deployment.Annotations["a-custom"] = "a-custom-value"
				delete(deployment.Labels, "l-3-")
				deployment.Labels["l-custom"] = "l-custom-value"
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents:  []string{"Normal AppsV1DeploymentUpdated apps.v1.Deployment_default/test updated"},
		},
		{
			name: "deletes and creates the Deployment when selector is changed and still matches the old pods",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.Spec.Template.Labels = map[string]string{
						"foo": "bar",
						"bar": "foo",
					}
					deployment.Spec.Selector = &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
					}
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Labels = map[string]string{
					"foo": "bar",
					"bar": "foo",
				}
				deployment.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				}
				return deployment
			}(),
			expectedDeployment: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.ResourceVersion = ""
				deployment.Spec.Template.Labels = map[string]string{
					"foo": "bar",
					"bar": "foo",
				}
				deployment.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				}
				utilruntime.Must(annotator.SetHashAnnotationWithCleanup(deployment))
				return deployment
			}(),
			expectedChanged: true,
			expectedErr:     nil,
			expectedEvents: []string{
				"Normal AppsV1DeploymentDeleted apps.v1.Deployment_default/test deleted",
				"Normal AppsV1DeploymentCreated apps.v1.Deployment_default/test created",
			},
		},
		{
			name: "apply fails when Deployment selector differs and existing Pod labels doesn't match new selector",
			existing: []runtime.Object{
				func() *appsv1.Deployment {
					deployment := newDeployment()
					deployment.Spec.Template.Labels = map[string]string{
						"foo": "bar",
					}
					deployment.Spec.Selector = &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"foo": "bar",
						},
					}
					return deployment
				}(),
			},
			required: func() *appsv1.Deployment {
				deployment := newDeployment()
				deployment.Spec.Template.Labels = map[string]string{
					"foo": "bar",
					"bar": "foo",
				}
				deployment.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
						"bar": "foo",
					},
				}
				return deployment
			}(),
			expectedDeployment: nil,
			expectedChanged:    false,
			expectedErr:        fmt.Errorf("can't get recreate reason: %w", fmt.Errorf(`required Deployment selector "bar=foo,foo=bar" doesn't match existing Pod Labels set map[foo:bar]`)),
			expectedEvents:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Client holds the state, so it has to persist the iterations.
			client := fake.NewClientset(tc.existing...)

			// ApplyDeployment needs to be reentrant so running it the second time should give the same results.
			// (One of the common mistakes is editing the object after computing the hash so it differs the second time.)
			iterations := 2
			if tc.expectedErr != nil {
				iterations = 1
			}
			for iteration := range iterations {
				t.Run("", func(t *testing.T) {
					ctx := t.Context()

					recorder := record.NewFakeRecorder(10)

					applier := resourceapply.NewApplier("internal.k8s-controller-lib/managed-hash", scheme, recorder)

					deploymentCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
					deploymentLister := appsv1listers.NewDeploymentLister(deploymentCache)

					if tc.cache != nil {
						for _, obj := range tc.cache {
							err := deploymentCache.Add(obj)
							if err != nil {
								t.Fatal(err)
							}
						}
					} else {
						deploymentList, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
							LabelSelector: labels.Everything().String(),
						})
						if err != nil {
							t.Fatal(err)
						}

						for i := range deploymentList.Items {
							err := deploymentCache.Add(&deploymentList.Items[i])
							if err != nil {
								t.Fatal(err)
							}
						}
					}

					gotDeployment, gotChanged, gotErr := ApplyDeployment(
						ctx,
						applier,
						client.AppsV1(),
						deploymentLister,
						tc.required,
						resourceapply.ApplyOptions{
							ForceOwnership: tc.forceOwnership,
						},
					)
					if !reflect.DeepEqual(gotErr, tc.expectedErr) {
						t.Fatalf("expected %v, got %v", tc.expectedErr, gotErr)
					}

					sanitize(gotDeployment)
					if !equality.Semantic.DeepEqual(gotDeployment, tc.expectedDeployment) {
						t.Errorf("expected %#v, got %#v, diff:\n%s", tc.expectedDeployment, gotDeployment, cmp.Diff(tc.expectedDeployment, gotDeployment))
					}

					// Make sure such object was actually created.
					if gotDeployment != nil {
						createdDeployment, err := client.AppsV1().Deployments(gotDeployment.Namespace).Get(ctx, gotDeployment.Name, metav1.GetOptions{})
						if err != nil {
							t.Error(err)
						}
						sanitize(createdDeployment)
						if !equality.Semantic.DeepEqual(createdDeployment, gotDeployment) {
							t.Errorf("created and returned deployments differ:\n%s", cmp.Diff(createdDeployment, gotDeployment))
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
							t.Errorf("expected %v, got %v, diff:\n%s", tc.expectedEvents, gotEvents, cmp.Diff(tc.expectedEvents, gotEvents))
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
