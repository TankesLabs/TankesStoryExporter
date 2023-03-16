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
	"github.com/omer-dayan/tankes-exporter/pkg/redisQuerier"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/omer-dayan/tankes-exporter/api/v1alpha1"
)

// TankesRedisMetricReconciler reconciles a TankesRedisMetric object
type TankesRedisMetricReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	*redisQuerier.Handler
}

//+kubebuilder:rbac:groups=core.tankes.story,resources=tankesredismetrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tankes.story,resources=tankesredismetrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.tankes.story,resources=tankesredismetrics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TankesRedisMetric object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TankesRedisMetricReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	fullyQualifiedName := req.NamespacedName.String()

	tankesRedisMetric := &corev1alpha1.TankesRedisMetric{}
	err := r.Client.Get(ctx, req.NamespacedName, tankesRedisMetric, &client.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			metricsHolder.Unregister(fullyQualifiedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if metricsHolder.IsRegistered(fullyQualifiedName) {
		promMetric := metricsHolder.GetCollectale(fullyQualifiedName).GetMetric()
		metricsHolder.UpdateCollectable(fullyQualifiedName, &redisCollectable{
			tankes:     tankesRedisMetric,
			prometheus: promMetric,
			handler:    r.Handler,
		})
		return ctrl.Result{}, nil
	}

	promMetric := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      tankesRedisMetric.Spec.Name,
		},
		tankesRedisMetric.Spec.LabelMatchingGroups,
	)
	metricsHolder.Register(fullyQualifiedName, &redisCollectable{
		tankes:     tankesRedisMetric,
		prometheus: promMetric,
		handler:    r.Handler,
	})

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TankesRedisMetricReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.TankesRedisMetric{}).
		Complete(r)
}

type redisCollectable struct {
	handler    *redisQuerier.Handler
	tankes     *corev1alpha1.TankesRedisMetric
	prometheus *prometheus.GaugeVec
}

func (r *redisCollectable) GetMetric() *prometheus.GaugeVec {
	return r.prometheus
}

func (r *redisCollectable) Collect(_ context.Context) error {
	return r.handler.CollectTankesRedisMetric(r.tankes, r.prometheus)
}
