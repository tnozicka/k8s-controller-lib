package queuehelpers

import (
	"k8s.io/client-go/tools/cache"
)

func UnwrapTombstoneIfPresent(obj any) any {
	tombstone, isTombstone := obj.(cache.DeletedFinalStateUnknown)
	if isTombstone {
		return tombstone.Obj
	}

	return obj
}
