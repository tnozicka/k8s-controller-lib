package cacmreconciler

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/examples/pkg/scheme"
	"github.com/tnozicka/k8s-controller-lib/pkg/controllerhelpers"
	"github.com/tnozicka/k8s-controller-lib/pkg/controllertools"
	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
	kuberesourceapply "github.com/tnozicka/k8s-controller-lib/pkg/resourceapply/kubernetes"
)

type CT = *corev1.Secret

var (
	controllerGVK = corev1.SchemeGroupVersion.WithKind("Secret")
)

func managedLabelsForSecret(s *corev1.Secret) map[string]string {
	return map[string]string{
		// This should match your label selector. In our case it's implicit 1:1 mapping.
		"k8s-controller-lib.tnozicka.github.io/managed-by": s.Name,
	}
}

func (c *Controller) sync(ctx context.Context, key types.NamespacedName) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing Secret", "Ref", key)
	defer func() {
		klog.V(4).InfoS("Finished syncing Secret", "Ref", key, "Duration", time.Since(startTime))
	}()

	s, err := c.secretLister.Secrets(key.Namespace).Get(key.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).InfoS("Secret doesn't exist", "Ref", key)
			return nil
		}

		return fmt.Errorf("can't get secret %q from cache: %w", key, err)
	}

	managedLabels := managedLabelsForSecret(s)
	controlleeSelector := labels.SelectorFromSet(managedLabels)

	var objectErrs []error

	cmMap, err := controllerhelpers.GetObjects[CT, *corev1.ConfigMap](
		ctx,
		s,
		controllerGVK,
		controlleeSelector,
		controllerhelpers.GetObjectsFuncs[CT, *corev1.ConfigMap]{
			CRMControl: controllertools.CRMControl[CT, *corev1.ConfigMap]{
				GetControllerUncachedFunc: c.KubeClient.CoreV1().Secrets(s.Namespace).Get,
				PatchObjectFunc:           c.KubeClient.CoreV1().ConfigMaps(s.Namespace).Patch,
			},
			ListObjectsFunc: c.configMapLister.List,
		},
		scheme.Scheme,
	)
	if err != nil {
		objectErrs = append(objectErrs, err)
	}

	objectErr := errors.Join(objectErrs...)
	if objectErr != nil {
		return objectErr
	}

	// TODO: This is where a status would be sync. We'd need to encode it into an annotation for Secrets.
	// status := s.Status.DeepCopy()
	// status.ObservedGeneration = s.Generation
	//
	// if s.DeletionTimestamp != nil {
	// 	return s.updateStatus(ctx, cs, status)
	// }

	tlsCert := s.Data[corev1.TLSCertKey]

	var requiredCMs []*corev1.ConfigMap

	if s.Annotations["k8s-controller-lib.tnozicka.github.io/sync"] == "true" {
		requiredCMs = append(requiredCMs, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: s.Name,
			},
			Data: map[string]string{
				corev1.TLSCertKey: string(tlsCert),
			},
		})
	}

	err = controllerhelpers.MintManagedDataToSlice(requiredCMs, s.Namespace, managedLabels, s, controllerGVK)
	if err != nil {
		return fmt.Errorf("can't mint managed data: %w", err)
	}

	var conditions []metav1.Condition
	err = controllerhelpers.PruneObjects(
		ctx,
		c.KubeClient.CoreV1().ConfigMaps(s.Namespace),
		c.EventRecorder,
		requiredCMs,
		cmMap,
		s.Generation,
		&conditions,
		scheme.Scheme,
	)
	if err != nil {
		return fmt.Errorf("can't prune objects: %w", err)
	}

	for _, cm := range requiredCMs {
		_, _, err = kuberesourceapply.ApplyConfigMap(
			ctx,
			c.ResourceApplier,
			c.KubeClient.CoreV1(),
			c.configMapLister,
			cm,
			resourceapply.ApplyOptions{},
		)
		if err != nil {
			return fmt.Errorf("can't apply ConfigMap %q: %w", cm.Name, err)
		}
	}

	return nil
}
