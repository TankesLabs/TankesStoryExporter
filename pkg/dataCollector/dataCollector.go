package dataCollector

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/omer-dayan/tankes-exporter/pkg/metricsHolder"
)

func CollectData() {
	metricsHolder.Collect(context.Background())
}
