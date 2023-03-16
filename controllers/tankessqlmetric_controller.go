/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/omer-dayan/tankes-exporter/pkg/metricsHolder"
	"github.com/omer-dayan/tankes-exporter/pkg/mysqlQuerier"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/omer-dayan/tankes-exporter/api/v1alpha1"
)

const subsystem = "tankes"

// TankesSqlMetricReconciler reconciles a TankesSqlMetric object
type TankesSqlMetricReconciler struct {
	client.Client
	*runtime.Scheme
	*mysqlQuerier.Handler
}

//+kubebuilder:rbac:groups=core.tankes.story,resources=tankessqlmetrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tankes.story,resources=tankessqlmetrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.tankes.story,resources=tankessqlmetrics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TankesSqlMetric object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TankesSqlMetricReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fullyQualifiedName := req.NamespacedName.String()

	tankesSqlMetric := &corev1alpha1.TankesSqlMetric{}
	err := r.Client.Get(ctx, req.NamespacedName, tankesSqlMetric, &client.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			metricsHolder.Unregister(fullyQualifiedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if metricsHolder.IsRegistered(fullyQualifiedName) {
		if metricsHolder.IsRegistered(fullyQualifiedName) {
			promMetric := metricsHolder.GetCollectale(fullyQualifiedName).GetMetric()
			metricsHolder.UpdateCollectable(fullyQualifiedName, &sqlCollectable{
				tankes:     tankesSqlMetric,
				prometheus: promMetric,
				handler:    r.Handler,
			})
			return ctrl.Result{}, nil
		}
	}

	promMetric := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      tankesSqlMetric.Spec.Name,
		},
		tankesSqlMetric.Spec.LabelFields,
	)
	metricsHolder.Register(fullyQualifiedName, &sqlCollectable{
		handler:    r.Handler,
		tankes:     tankesSqlMetric,
		prometheus: promMetric,
	})
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TankesSqlMetricReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.TankesSqlMetric{}).
		Complete(r)
}

type sqlCollectable struct {
	handler    *mysqlQuerier.Handler
	tankes     *corev1alpha1.TankesSqlMetric
	prometheus *prometheus.GaugeVec
}

func (s *sqlCollectable) GetMetric() *prometheus.GaugeVec {
	return s.prometheus
}

func (s *sqlCollectable) Collect(_ context.Context) error {
	return s.handler.CollectTankesSqlMetric(s.tankes, s.prometheus)
}
