package controllerhelpers

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func IsDeploymentRolledOut(d *appsv1.Deployment) (bool, error) {
	if d.Spec.Replicas == nil {
		// This should never happen, but better safe than sorry.
		return false, fmt.Errorf("spec.replicas can't be nil")
	}

	if d.Status.ObservedGeneration == 0 || d.Generation > d.Status.ObservedGeneration {
		return false, nil
	}

	if d.Status.Replicas != *d.Spec.Replicas {
		return false, nil
	}

	if d.Status.ReadyReplicas < *d.Spec.Replicas {
		return false, nil
	}

	if d.Status.AvailableReplicas < *d.Spec.Replicas {
		return false, nil
	}

	return true, nil
}
