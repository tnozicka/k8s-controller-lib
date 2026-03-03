package healthz

import (
	"fmt"
	"net/http"

	"github.com/tnozicka/k8s-controller-lib/pkg/informerfactory"
)

type UnnamedHealthChecker interface {
	Check(req *http.Request) error
}

type HealthCheckers []UnnamedHealthChecker

func (cs HealthCheckers) Check(req *http.Request) error {
	for _, c := range cs {
		err := c.Check(req)
		if err != nil {
			return err
		}
	}
	return nil
}

type InformerSyncCheck struct {
	informerfactory.StartedInformersGetter
}

func (i *InformerSyncCheck) Check(_ *http.Request) error {
	var notStarted []string
	for id, started := range i.GetStartedInformerMap() {
		if !started {
			notStarted = append(notStarted, id)
		}
	}

	notStartedCount := len(notStarted)
	if notStartedCount > 0 {
		return fmt.Errorf("%d informers not started yet: %v", notStartedCount, notStarted)
	}

	return nil
}
