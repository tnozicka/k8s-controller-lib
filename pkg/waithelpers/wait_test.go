package waithelpers

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

const (
	testNamespaceName           = "test-ns"
	testAnnotationKey           = "test-annotation"
	testUID           types.UID = "test-uid"
)

func hasValueCond[T kubetypes.Object](t *testing.T, expected string) func(obj T) (bool, error) {
	return func(obj T) (bool, error) {
		var actual string
		if obj.GetAnnotations() != nil {
			actual = obj.GetAnnotations()[testAnnotationKey]
		}
		res := actual == expected
		t.Logf("hasValueCond: done=%t, actual=%q, expected=%q", res, actual, expected)
		return res, nil
	}
}

func newTestMeta(value string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      "test-cm",
		Namespace: testNamespaceName,
		UID:       testUID,
		Annotations: map[string]string{
			testAnnotationKey: value,
		},
	}
}

func newTestMetaClusterScoped(value string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: "test-cm",
		UID:  testUID,
		Annotations: map[string]string{
			testAnnotationKey: value,
		},
	}
}

func testWaitForState[Object kubetypes.Object, Client any](
	t *testing.T,
	client Client,
	name string,
	waitFunc func(
		context.Context,
		Client,
		string,
		WaitForStateOptions,
		...func(Object) (bool, error),
	) (Object, error),
) {
	t.Helper()

	_, err := waitFunc(
		t.Context(),
		client,
		name,
		WaitForStateOptions{},
		hasValueCond[Object](t, "any"),
	)
	if err != nil {
		t.Errorf("can't wait for state: %v", err)
	}
}

func deleteObject[Client interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}](ctx context.Context, client Client, name string) error {
	return client.Delete(ctx, name, metav1.DeleteOptions{})
}

func testWaitForDeletion[Object kubetypes.Object, Client any](
	t *testing.T,
	typeName string,
	obj Object,
	client Client,
	deleteFunc func(ctx context.Context, client Client, name string) error,
	waitFunc func(context.Context, Client, types.NamespacedName, *types.UID) error,
) {
	t.Helper()

	synctest.Test(t, func(t *testing.T) {
		cycleWaitTimeout := 24 * time.Hour

		var wg wait.Group
		defer wg.Wait()

		waitCtx, waitCtxCancel := context.WithTimeout(t.Context(), 2*cycleWaitTimeout)
		defer waitCtxCancel()

		wg.StartWithContext(waitCtx, func(ctx context.Context) {
			time.Sleep(cycleWaitTimeout)
			synctest.Wait()
			err := deleteFunc(ctx, client, obj.GetName())
			if err != nil {
				t.Fatalf("can't delete %s: %v", typeName, err)
			}
		})

		uid := obj.GetUID()
		err := waitFunc(
			waitCtx,
			client,
			naming.ObjNN(obj),
			&uid,
		)
		if err != nil {
			t.Errorf("can't wait for %s deletion: %v", typeName, err)
		}
	})
}

func TestWaitForObjectState(t *testing.T) {
	t.Parallel()

	newTestCM := func(value string) *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: newTestMeta(value),
		}
	}
	testCMNN := naming.ObjNN(newTestCM(""))

	hasValueCond := hasValueCond[*corev1.ConfigMap]

	// cycleWaitTimeout is the longest time it shall take WaitForObjectState to observe the current state.
	// Now that go has synctest this can be set artificially high
	// so it doesn't fail on interactive debugging sessions or flakes.
	cycleWaitTimeout := 24 * time.Hour

	tt := []struct {
		name        string
		object      kubetypes.Object
		options     WaitForStateOptions
		conditions  []func(*corev1.ConfigMap) (bool, error)
		modifyFunc  func(context.Context, *testing.T, corev1client.ConfigMapInterface)
		timeout     time.Duration
		expectedObj *corev1.ConfigMap
		expectedErr error
	}{
		{
			name:    "added event with condition met",
			object:  newTestCM("initial"),
			options: WaitForStateOptions{},
			conditions: []func(*corev1.ConfigMap) (bool, error){
				hasValueCond(t, "initial"),
				hasValueCond(t, "initial"),
			},
			modifyFunc:  nil,
			timeout:     cycleWaitTimeout,
			expectedObj: newTestCM("initial"),
			expectedErr: nil,
		},
		{
			name:    "added event with condition unmet",
			object:  newTestCM("initial"),
			options: WaitForStateOptions{},
			conditions: []func(*corev1.ConfigMap) (bool, error){
				hasValueCond(t, "initial"),
				hasValueCond(t, "modified"),
			},
			modifyFunc:  nil,
			timeout:     cycleWaitTimeout,
			expectedObj: nil,
			expectedErr: wait.ErrorInterrupted(nil),
		},
		{
			name:    "modify event with condition met",
			object:  newTestCM("initial"),
			options: WaitForStateOptions{},
			conditions: []func(*corev1.ConfigMap) (bool, error){
				hasValueCond(t, "modified"),
			},
			modifyFunc: func(ctx context.Context, t *testing.T, client corev1client.ConfigMapInterface) {
				time.Sleep(cycleWaitTimeout)
				synctest.Wait()
				cm := newTestCM("modified")
				_, err := client.Update(ctx, cm, metav1.UpdateOptions{})
				if err != nil {
					t.Fatalf("can't update ConfigMap %q: %v", naming.ObjNN(cm), err)
				}
				t.Logf("updated ConfigMap %s/%s", cm.Namespace, cm.Name)
			},
			timeout:     2 * cycleWaitTimeout,
			expectedObj: newTestCM("modified"),
			expectedErr: nil,
		},
		{
			name:    "modify event with condition unmet",
			object:  newTestCM("initial"),
			options: WaitForStateOptions{},
			conditions: []func(*corev1.ConfigMap) (bool, error){
				hasValueCond(t, "different"),
			},
			modifyFunc: func(ctx context.Context, t *testing.T, client corev1client.ConfigMapInterface) {
				time.Sleep(cycleWaitTimeout)
				synctest.Wait()
				cm := newTestCM("modified")
				_, err := client.Update(ctx, cm, metav1.UpdateOptions{})
				if err != nil {
					t.Fatalf("can't update ConfigMap %q: %v", naming.ObjNN(cm), err)
				}
				t.Logf("updated ConfigMap %q", testCMNN)
			},
			timeout:     cycleWaitTimeout,
			expectedObj: nil,
			expectedErr: wait.ErrorInterrupted(nil),
		},
		{
			name:    "delete event with condition unmet",
			object:  newTestCM("initial"),
			options: WaitForStateOptions{},
			conditions: []func(*corev1.ConfigMap) (bool, error){
				hasValueCond(t, "different"),
			},
			modifyFunc: func(ctx context.Context, t *testing.T, client corev1client.ConfigMapInterface) {
				time.Sleep(cycleWaitTimeout)
				synctest.Wait()
				err := client.Delete(ctx, testCMNN.Name, metav1.DeleteOptions{})
				if err != nil {
					t.Fatalf("can't delete ConfigMap %q: %v", testCMNN, err)
				}
				t.Logf("deleted ConfigMap %q", testCMNN)
			},
			timeout:     2 * cycleWaitTimeout,
			expectedObj: nil,
			expectedErr: fmt.Errorf("unexpected event type %q", watch.Deleted),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			synctest.Test(t, func(t *testing.T) {
				cmClient := fake.NewClientset(tc.object).CoreV1().ConfigMaps(testNamespaceName)

				var wg wait.Group
				defer wg.Wait()

				waitCtx, waitCtxCancel := context.WithTimeout(t.Context(), tc.timeout)
				defer waitCtxCancel()

				if tc.modifyFunc != nil {
					wg.StartWithContext(waitCtx, func(ctx context.Context) {
						tc.modifyFunc(ctx, t, cmClient)
					})
				}

				got, gotErr := WaitForObjectState[*corev1.ConfigMap, *corev1.ConfigMapList](
					waitCtx,
					cmClient,
					tc.object.GetName(),
					tc.options,
					tc.conditions...,
				)
				if !reflect.DeepEqual(gotErr, tc.expectedErr) {
					t.Errorf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, gotErr, cmpopts.EquateErrors()))
				}

				if got != nil {
					got.SetManagedFields(nil)
				}
				if !apiequality.Semantic.DeepEqual(got, tc.expectedObj) {
					t.Errorf("expected and got objects differ:\n%s", cmp.Diff(tc.expectedObj, got))
				}
			})
		})
	}
}

func TestWaitForObjectDeletion(t *testing.T) {
	t.Parallel()

	newTestCM := func() *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: newTestMeta(""),
		}
	}
	testCMNN := naming.ObjNN(newTestCM())

	cycleWaitTimeout := 24 * time.Hour

	tt := []struct {
		name        string
		object      *corev1.ConfigMap
		modifyFunc  func(context.Context, *testing.T, corev1client.ConfigMapInterface)
		timeout     time.Duration
		expectedErr error
	}{
		{
			name:        "object is already deleted",
			object:      nil,
			modifyFunc:  nil,
			timeout:     cycleWaitTimeout,
			expectedErr: nil,
		},
		{
			name:   "existing object gets deleted",
			object: newTestCM(),
			modifyFunc: func(ctx context.Context, t *testing.T, client corev1client.ConfigMapInterface) {
				time.Sleep(cycleWaitTimeout)
				synctest.Wait()
				err := client.Delete(ctx, testCMNN.Name, metav1.DeleteOptions{})
				if err != nil {
					t.Fatalf("can't delete ConfigMap: %v", err)
				}
			},
			timeout:     2 * cycleWaitTimeout,
			expectedErr: nil,
		},
		{
			name:        "object never gets deleted",
			object:      newTestCM(),
			modifyFunc:  nil,
			timeout:     cycleWaitTimeout,
			expectedErr: wait.ErrorInterrupted(nil),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			synctest.Test(t, func(t *testing.T) {
				var err error

				kubeClient := fake.NewClientset()
				if tc.object != nil {
					err = kubeClient.Tracker().Add(tc.object)
					if err != nil {
						t.Fatalf("can't add object into the tracker: %v", err)
					}
				}
				cmClient := kubeClient.CoreV1().ConfigMaps(testNamespaceName)

				var wg wait.Group
				defer wg.Wait()

				waitCtx, waitCtxCancel := context.WithTimeout(t.Context(), tc.timeout)
				defer waitCtxCancel()

				if tc.modifyFunc != nil {
					wg.StartWithContext(waitCtx, func(ctx context.Context) {
						tc.modifyFunc(ctx, t, cmClient)
					})
				}

				var uid *types.UID
				if tc.object != nil {
					uid = &tc.object.UID
				}
				gotErr := WaitForObjectDeletion[*corev1.ConfigMap, *corev1.ConfigMapList](
					waitCtx,
					cmClient,
					testCMNN,
					uid,
				)
				if !reflect.DeepEqual(gotErr, tc.expectedErr) {
					t.Errorf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, gotErr, cmpopts.EquateErrors()))
				}
			})
		})
	}
}
