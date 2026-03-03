package assets

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type Bytes[T runtime.Object] []byte

func (b Bytes[T]) Decode(cf serializer.CodecFactory) (T, error) {
	obj := reflect.New(reflect.TypeOf(*new(T)).Elem()).Interface().(T)

	err := runtime.DecodeInto(cf.UniversalDeserializer(), b, obj)
	if err != nil {
		return obj, err
	}

	return obj, nil
}
