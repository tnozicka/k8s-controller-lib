package controllerhelpers

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
)

func TestGVKAsTitleCaseString(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		gvk            schema.GroupVersionKind
		expectedString string
	}{
		{
			name: "all fields populated",
			gvk: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			expectedString: "AppsV1Deployment",
		},
		{
			name: "core group",
			gvk: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			expectedString: "V1Pod",
		},
		{
			name: "multiword Kind",
			gvk: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1beta1",
				Kind:    "CronJob",
			},
			expectedString: "BatchV1beta1Cronjob",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := gvkAsTitleCase(tc.gvk).String()
			if !reflect.DeepEqual(got, tc.expectedString) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedString, got))
			}
		})
	}
}

func TestReportEvent(t *testing.T) {
	t.Parallel()

	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	}

	obj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-config",
			Namespace: "default",
		},
	}

	tt := []struct {
		name           string
		operationErr   error
		verb           string
		expectedEvents []string
	}{
		{
			name:           "successful creation event",
			operationErr:   nil,
			verb:           "create",
			expectedEvents: []string{"Normal V1ConfigmapCreated v1.ConfigMap_default/my-config created"},
		},
		{
			name:           "successful deletion event",
			operationErr:   nil,
			verb:           "delete",
			expectedEvents: []string{"Normal V1ConfigmapDeleted v1.ConfigMap_default/my-config deleted"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := record.NewFakeRecorder(1 + 1)
			ReportEvent(recorder, obj, gvk, tc.operationErr, tc.verb)

			close(recorder.Events)
			var events []string
			for e := range recorder.Events {
				events = append(events, e)
			}

			if !reflect.DeepEqual(events, tc.expectedEvents) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedEvents, events))
			}
		})
	}
}
