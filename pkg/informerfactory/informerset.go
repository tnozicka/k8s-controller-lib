package informerfactory

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/helpers/maps"
)

type NewInformerFunc func(InformerFactory) kubeinformers.GenericInformer

type InformerEntry struct {
	GVR         schema.GroupVersionResource
	NewInformer NewInformerFunc
}

type InformerSetFactory struct {
	lock             sync.Mutex
	informerBuilders map[schema.GroupVersionResource]NewInformerFunc
	informers        map[schema.GroupVersionResource]kubeinformers.GenericInformer
	startedInformers sets.Set[schema.GroupVersionResource]

	shuttingDown bool
	wg           wait.Group
}

var _ InformerFactory = &InformerSetFactory{}

func NewInformerSetFactory(informers ...InformerEntry) (*InformerSetFactory, error) {
	is := &InformerSetFactory{
		informerBuilders: make(map[schema.GroupVersionResource]NewInformerFunc, len(informers)),
		informers:        map[schema.GroupVersionResource]kubeinformers.GenericInformer{},
		startedInformers: sets.New[schema.GroupVersionResource](),
	}

	for _, informerEntry := range informers {
		_, exists := is.informerBuilders[informerEntry.GVR]
		if exists {
			return nil, fmt.Errorf("informer for %s is already present in the set", informerEntry.GVR)
		}

		is.informerBuilders[informerEntry.GVR] = informerEntry.NewInformer
	}

	return is, nil
}

// Start starts all informers that have been created and not started yet.
func (is *InformerSetFactory) Start(stopCh <-chan struct{}) {
	is.lock.Lock()
	defer is.lock.Unlock()

	select {
	case <-stopCh:
		klog.V(2).InfoS("Skipping informer set factory start because stop channel is already closed")
		return
	default:
	}

	if is.shuttingDown {
		klog.V(2).InfoS("Skipping informer set factory start because it is shutting down")
		return
	}

	for gvr, informer := range is.informers {
		if is.startedInformers.Has(gvr) {
			continue
		}

		is.wg.StartWithChannel(stopCh, informer.Informer().Run)
		is.startedInformers.Insert(gvr)
	}
}

func (is *InformerSetFactory) Shutdown() {
	defer is.wg.Wait()

	is.lock.Lock()
	is.shuttingDown = true
	is.lock.Unlock()
}

func (is *InformerSetFactory) ForResource(gvr schema.GroupVersionResource) (kubeinformers.GenericInformer, error) {
	is.lock.Lock()
	defer is.lock.Unlock()

	informer, informerExists := is.informers[gvr]
	if informerExists {
		return informer, nil
	}

	informerBuilder, informerBuilderExists := is.informerBuilders[gvr]
	if !informerBuilderExists {
		return nil, fmt.Errorf("informer for %q does not exist", gvr)
	}

	informer = informerBuilder(is)
	is.informers[gvr] = informer

	return informer, nil
}

func (is *InformerSetFactory) GetStartedInformerMap() map[string]bool {
	is.lock.Lock()
	defer is.lock.Unlock()

	res := map[string]bool{}

	for gvr := range is.informers {
		res[gvr.String()] = maps.HasKey(is.startedInformers, gvr)
	}

	return res
}
