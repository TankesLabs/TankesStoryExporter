package redisQuerier

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/omer-dayan/tankes-exporter/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strconv"
)

type Handler struct {
	client *redis.Client
}

func New(server string) (*Handler, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     server,
		Password: "",
		DB:       0,
	})
	return &Handler{
		client: client,
	}, client.Ping().Err()
}

func (h *Handler) Close() error {
	return h.client.Close()
}

func (h *Handler) CollectTankesRedisMetric(tankes *v1alpha1.TankesRedisMetric, prom *prometheus.GaugeVec) error {
	keys, err := h.client.Keys("*").Result()
	if err != nil {
		return err
	}

	re := regexp.MustCompile(tankes.Spec.Regex)
	for _, key := range keys {
		match := re.FindStringSubmatch(key)
		if len(match) > 0 {
			labelValues := match[1:]
			if len(labelValues) != len(tankes.Spec.LabelMatchingGroups) {
				return fmt.Errorf("TankesRedisMetric had different number of matching groups than labels")
			}
			valueStr, err := h.client.Get(key).Result()
			if err != nil {
				return err
			}
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				value = 0
			}
			prom.WithLabelValues(labelValues...).Set(value)
		}
	}
	return nil
}
