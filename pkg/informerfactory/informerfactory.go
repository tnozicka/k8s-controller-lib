package informerfactory

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicinformers "k8s.io/client-go/dynamic/dynamicinformer"
	kubeinformers "k8s.io/client-go/informers"

	"github.com/tnozicka/k8s-controller-lib/pkg/helpers/maps"
)

type GenericForResourceFunc func(schema.GroupVersionResource) (kubeinformers.GenericInformer, error)

type StartedInformersGetter interface {
	GetStartedInformerMap() map[string]bool
}

type InformerFactoryBase interface {
	Start(stopCh <-chan struct{})
	Shutdown()
}

type InformerFactoryBaseWithCacheWaiter interface {
	InformerFactoryBase
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}

type InformerFactory interface {
	InformerFactoryBase
	ForResource(resource schema.GroupVersionResource) (kubeinformers.GenericInformer, error)
	StartedInformersGetter
}

type factoryAdapter[G any] struct {
	InformerFactoryBaseWithCacheWaiter
	forResourceFN func(schema.GroupVersionResource) (G, error)
}

var _ InformerFactory = &factoryAdapter[any]{}

func NewFactoryAdapter[G any](factory InformerFactoryBaseWithCacheWaiter, forResourceFunc func(schema.GroupVersionResource) (G, error)) InformerFactory {
	return &factoryAdapter[G]{
		InformerFactoryBaseWithCacheWaiter: factory,
		forResourceFN:                      forResourceFunc,
	}
}

func (fa *factoryAdapter[G]) ForResource(gvr schema.GroupVersionResource) (kubeinformers.GenericInformer, error) {
	g, err := fa.forResourceFN(gvr)
	gi, ok := any(g).(kubeinformers.GenericInformer)
	if !ok {
		return nil, fmt.Errorf("%s is not a kubeinformers.GenericInformer", gvr.String())
	}
	return gi, err
}

func (fa *factoryAdapter[G]) GetStartedInformerMap() map[string]bool {
	closedChan := make(chan struct{})
	close(closedChan)
	m := fa.WaitForCacheSync(closedChan)

	res := map[string]bool{}
	for t, synced := range m {
		res[t.String()] = synced
	}
	return res
}

// informerFactory groups all typed informers and fallbacks to using dynamic informers for any other resource.
type informerFactory struct {
	typedInformerFactories []InformerFactory
	dynamicInformerFactory dynamicinformers.DynamicSharedInformerFactory
}

var _ InformerFactory = &informerFactory{}

func NewInformerFactory(
	typedInformerFactories []InformerFactory,
	dynamicInformerFactory dynamicinformers.DynamicSharedInformerFactory,
) InformerFactory {
	return &informerFactory{
		typedInformerFactories: typedInformerFactories,
		dynamicInformerFactory: dynamicInformerFactory,
	}
}

func (i *informerFactory) ForResource(resource schema.GroupVersionResource) (kubeinformers.GenericInformer, error) {
	var lookupErr error
	for _, factory := range i.typedInformerFactories {
		var informer kubeinformers.GenericInformer
		informer, lookupErr = factory.ForResource(resource)
		if lookupErr == nil {
			return informer, nil
		}
	}
	return i.dynamicInformerFactory.ForResource(resource), nil
}

func (i *informerFactory) Start(stopCh <-chan struct{}) {
	for _, factory := range i.typedInformerFactories {
		factory.Start(stopCh)
	}
	i.dynamicInformerFactory.Start(stopCh)
}

func (i *informerFactory) Shutdown() {
	for _, factory := range i.typedInformerFactories {
		factory.Shutdown()
	}
	i.dynamicInformerFactory.Shutdown()
}

func (i *informerFactory) GetStartedInformerMap() map[string]bool {
	res := map[string]bool{}

	for _, tf := range i.typedInformerFactories {
		m := tf.GetStartedInformerMap()
		for id, synced := range m {
			if maps.HasKey(res, id) {
				panic(fmt.Sprintf("conflicting type informer id %q", id))
			}
			res[id] = synced
		}
	}

	closedChan := make(chan struct{})
	close(closedChan)
	gvrMap := i.dynamicInformerFactory.WaitForCacheSync(closedChan)
	for gvr, synced := range gvrMap {
		id := gvr.String()
		if maps.HasKey(res, id) {
			panic(fmt.Sprintf("conflicting type informer id %q", id))
		}
		res[id] = synced
	}

	return res
}
