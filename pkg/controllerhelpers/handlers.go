package controllerhelpers

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
)

func toTypedObj[T kubetypes.Object](untypedObj any) T {
	obj, ok := untypedObj.(T)
	if !ok {
		runtime.HandleError(fmt.Errorf("invalid object type: expected object %T, not a %T", obj, untypedObj))
	}

	return obj
}

func ToUntypedHandler[T kubetypes.Object](f func(T)) func(any) {
	return func(untypedObj any) {
		obj := toTypedObj[T](untypedObj)
		f(obj)
	}
}

func DisregardOldObjectHandler(handler func(obj any)) func(any, any) {
	return func(_, cur any) {
		handler(cur)
	}
}

func GetUnifiedResourceHandlersUntyped(handler func(obj any)) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    handler,
		UpdateFunc: DisregardOldObjectHandler(handler),
		DeleteFunc: handler,
	}
}

func GetUnifiedResourceHandlers[T kubetypes.Object](handler func(obj T)) cache.ResourceEventHandlerFuncs {
	return GetUnifiedResourceHandlersUntyped(ToUntypedHandler[T](handler))
}
