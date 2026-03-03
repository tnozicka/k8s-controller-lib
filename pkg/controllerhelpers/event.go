package controllerhelpers

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type gvkAsTitleCase schema.GroupVersionKind

func (gvk gvkAsTitleCase) String() string {
	titleCaser := cases.Title(language.AmericanEnglish)
	return fmt.Sprintf(
		"%s%s%s",
		titleCaser.String(gvk.Group),
		titleCaser.String(gvk.Version),
		titleCaser.String(gvk.Kind),
	)
}

func ReportEvent(recorder record.EventRecorder, obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error, verb string) {
	titleCaser := cases.Title(language.AmericanEnglish)
	verbTitleCase := titleCaser.String(verb)
	gvkTitleCase := gvkAsTitleCase(gvk)

	if operationErr != nil {
		recorder.Eventf(
			obj,
			corev1.EventTypeWarning,
			fmt.Sprintf(
				"%s%sFailed",
				verbTitleCase,
				gvkTitleCase,
			),
			"Failed to %s %s: %v",
			verb,
			naming.ObjKindNN(gvk, obj),
			operationErr,
		)
		return
	}

	recorder.Eventf(
		obj,
		corev1.EventTypeNormal,
		fmt.Sprintf(
			"%s%sd",
			gvkTitleCase,
			verbTitleCase,
		),
		"%s %sd",
		naming.ObjKindNN(gvk, obj),
		verb,
	)
}

func ReportCreateEvent(recorder record.EventRecorder, obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error) {
	if apierrors.HasStatusCause(operationErr, corev1.NamespaceTerminatingCause) {
		klog.V(4).InfoS("Skipping event reporting because the namespace is terminating", "ObjRef", klog.KObj(obj), "GVK", gvk)
		// If the namespace is being terminated,
		// we can safely ignore this error since all subsequent creations will fail.
		return
	}

	ReportEvent(recorder, obj, gvk, operationErr, "create")
}

func ReportUpdateEvent(recorder record.EventRecorder, obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error) {
	ReportEvent(recorder, obj, gvk, operationErr, "update")
}

func ReportDeleteEvent(recorder record.EventRecorder, obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error) {
	ReportEvent(recorder, obj, gvk, operationErr, "delete")
}
