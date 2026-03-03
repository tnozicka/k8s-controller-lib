package controllerhelpers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestIsDeploymentRolledOut(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		deployment     *appsv1.Deployment
		expectedResult bool
		expectedErr    error
	}{
		{
			name: "nil replicas returns error",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: nil,
				},
			},
			expectedResult: false,
			expectedErr:    fmt.Errorf("spec.replicas can't be nil"),
		},
		{
			name: "not rolled out when observedGeneration is lower",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 0,
					Replicas:           3,
					ReadyReplicas:      3,
					AvailableReplicas:  3,
				},
			},
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name: "not rolled out when status replicas is less than desired",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 1,
					Replicas:           2,
					ReadyReplicas:      3,
					AvailableReplicas:  3,
				},
			},
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name: "not rolled out when status ready replicas is less than desired",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 1,
					Replicas:           3,
					ReadyReplicas:      2,
					AvailableReplicas:  3,
				},
			},
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name: "not rolled out when status available replicas is less than desired",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 1,
					Replicas:           3,
					ReadyReplicas:      3,
					AvailableReplicas:  2,
				},
			},
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name: "fully rolled out",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 1,
					Replicas:           3,
					ReadyReplicas:      3,
					AvailableReplicas:  3,
				},
			},
			expectedResult: true,
			expectedErr:    nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := IsDeploymentRolledOut(tc.deployment)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}
			if got != tc.expectedResult {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedResult, got))
			}
		})
	}
}
