package metricsHolder

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sync"
)

var mutex = &sync.Mutex{}

var active = map[string]Collectable{}

type Collectable interface {
	Collect(context.Context) error
	GetMetric() *prometheus.GaugeVec
}

func Register(name string, collectable Collectable) {
	mutex.Lock()
	defer mutex.Unlock()

	active[name] = collectable
	metrics.Registry.MustRegister(collectable.GetMetric())
}

func Unregister(name string) {
	mutex.Lock()
	defer mutex.Unlock()

	metrics.Registry.Unregister(*active[name].GetMetric())
	delete(active, name)
}

func IsRegistered(name string) bool {
	_, exists := active[name]
	return exists
}

func GetCollectale(name string) Collectable {
	return active[name]
}

func UpdateCollectable(name string, collectable Collectable) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := active[name]; !exists {
		return fmt.Errorf("collectable %s does not exists on set", name)
	}
	active[name] = collectable
	return nil
}

func Collect(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	logger := log.FromContext(ctx)

	for name, collectable := range active {
		err := collectable.Collect(ctx)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Could not perform query on collectable %s", name))
		}
	}
}
