package cacmreconciler

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/controller"
	"github.com/tnozicka/k8s-controller-lib/pkg/controllerhelpers"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

type Controller struct {
	controller.Base[types.NamespacedName]

	secretLister    corev1listers.SecretLister
	configMapLister corev1listers.ConfigMapLister

	ResourceApplier *resourceapply.Applier
}

func NewController(
	scheme *runtime.Scheme,
	managedHashAnnotationKey string,
	kubeClient kubernetes.Interface,
	secretInformer corev1informers.SecretInformer,
	configMapInformer corev1informers.ConfigMapInformer,
) (*Controller, error) {
	c := &Controller{
		secretLister:    secretInformer.Lister(),
		configMapLister: configMapInformer.Lister(),
	}
	bc, err := controller.NewBase(
		"cacm-reconciler",
		scheme,
		kubeClient,
		[]cache.InformerSynced{
			secretInformer.Informer().HasSynced,
			configMapInformer.Informer().HasSynced,
		},
		controller.BaseControls[types.NamespacedName]{
			SyncFunc: c.sync,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create base controller: %w", err)
	}

	c.Base = *bc
	c.ResourceApplier = resourceapply.NewApplier(
		managedHashAnnotationKey,
		scheme,
		bc.EventRecorder,
	)

	_, err = secretInformer.Informer().AddEventHandler(
		controllerhelpers.GetUnifiedResourceHandlers(c.enqueue),
	)
	if err != nil {
		return nil, fmt.Errorf("can't add secret handlers: %w", err)
	}

	_, err = configMapInformer.Informer().AddEventHandler(
		controllerhelpers.GetUnifiedResourceHandlers(c.queueConfigMap),
	)
	if err != nil {
		return nil, fmt.Errorf("can't add config map handlers: %w", err)
	}

	return c, nil
}

func (c *Controller) enqueue(s *corev1.Secret) {
	c.Queue.Add(naming.ObjNN(s))
}

func (c *Controller) getSecretsForConfigMap(cm *corev1.ConfigMap) ([]*corev1.Secret, error) {
	if len(cm.Labels) == 0 {
		return nil, fmt.Errorf("can't find Secrets for ConfigMap %q because it has no labels", naming.ObjNN(cm))
	}

	// Normally, this is where you'd have to list all objects and enque all that could select this object.
	// In this example we don't have configurable selectors and use 1:1 mapping by name.

	s, err := c.secretLister.Secrets(cm.Namespace).Get(cm.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get Secret %q for ConfigMap %q: %w", types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name}, naming.ObjNN(cm), err)
	}

	return []*corev1.Secret{s}, nil
}

func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *corev1.Secret {
	if controllerRef.Kind != controllerGVK.Kind {
		return nil
	}

	s, err := c.secretLister.Secrets(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	if s.UID != controllerRef.UID {
		return nil
	}

	return s
}

func (c *Controller) queueConfigMap(cm *corev1.ConfigMap) {
	controllerRef := metav1.GetControllerOf(cm)
	if controllerRef != nil {
		s := c.resolveControllerRef(cm.Namespace, controllerRef)
		if s != nil {
			klog.V(4).InfoS("ConfigMap added", "ConfigMap", klog.KObj(cm))
			c.Queue.Add(naming.ObjNN(s))
			return
		}
	}

	secrets, err := c.getSecretsForConfigMap(cm)
	if err != nil {
		// In handlers, it's ok to skip objects that we don't control (and these errors).
		return
	}
	for _, s := range secrets {
		c.Queue.Add(naming.ObjNN(s))
	}
}
