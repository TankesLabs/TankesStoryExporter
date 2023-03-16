package mysqlQuerier

import (
	"database/sql"
	"github.com/omer-dayan/tankes-exporter/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

type Handler struct {
	Db *sql.DB
}

func New(connectionString string) (*Handler, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &Handler{
		Db: db,
	}, nil
}

func (h *Handler) Close() error {
	return h.Db.Close()
}

func (h *Handler) CollectTankesSqlMetric(tankes *v1alpha1.TankesSqlMetric, prom *prometheus.GaugeVec) error {
	rows, err := h.Db.Query(tankes.Spec.Query)
	if err != nil {
		return err
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		values := make([]string, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		err = rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}
		row := map[string]string{}
		for i, col := range columns {
			row[col] = values[i]
		}

		labelValues := []string{}
		for _, labelField := range tankes.Spec.LabelFields {
			labelValues = append(labelValues, row[labelField])
		}
		value, err := strconv.ParseFloat(row[tankes.Spec.ValueField], 64)
		if err != nil {
			return err
		}
		prom.WithLabelValues(labelValues...).Set(value)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
