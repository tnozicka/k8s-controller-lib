package queuehelpers

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func HandleError2[T any](v T, err error) T {
	utilruntime.HandleError(err)
	return v
}
