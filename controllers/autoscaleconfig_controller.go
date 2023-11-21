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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalerv1alpha1 "github.com/percona/pvc-autoscaler/api/v1alpha1"
	"github.com/prometheus/common/expfmt"
)

// AutoscaleConfigReconciler reconciles a AutoscaleConfig object
type AutoscaleConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=nodes/metrics,verbs=get;list;watch
// +kubebuilder:rbac:groups=autoscaler.percona.com,resources=autoscaleconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaler.percona.com,resources=autoscaleconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaler.percona.com,resources=autoscaleconfigs/finalizers,verbs=update

func (r *AutoscaleConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	rr := ctrl.Result{RequeueAfter: 30 * time.Second}

	cr := &autoscalerv1alpha1.AutoscaleConfig{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, cr)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return rr, err
	}

	f, err := os.Open("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return rr, err
	}

	token, err := io.ReadAll(f)
	if err != nil {
		return rr, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			// Disable TLS verification
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	nodeList := corev1.NodeList{}
	err = r.Client.List(ctx, &nodeList)
	if err != nil {
		return rr, err
	}

	threshold := cr.Spec.ThresholdValue.AsApproximateFloat64()
	l.Info("Read threshold value", "threshold", cr.Spec.ThresholdValue, "float64", threshold)

	for _, node := range nodeList.Items {
		url := fmt.Sprintf("https://%s:%d/metrics", node.Status.Addresses[0].Address, node.Status.DaemonEndpoints.KubeletEndpoint.Port)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return rr, err
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		if err != nil {
			return rr, err
		}
		defer resp.Body.Close()

		// Parse the Prometheus metrics
		parser := expfmt.TextParser{}
		metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
		if err != nil {
			return rr, err
		}

		for name, mf := range metricFamilies {
			if strings.Contains(name, "kubelet_volume_stats_available_bytes") {
				for _, m := range mf.Metric {
					l.Info("Read metric", "node", node.Name, "metric", name, "volume", m.Label[1].GetValue(), "value", m.Gauge.GetValue())
					if m.Gauge.GetValue() < threshold {
						l.Info("Scaling up", "node", node.Name, "volume", m.Label[1].GetValue(), "threshold", threshold, "value", m.Gauge.GetValue())
					}
				}
			}
		}
	}

	return rr, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoscaleConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalerv1alpha1.AutoscaleConfig{}).
		Complete(r)
}
