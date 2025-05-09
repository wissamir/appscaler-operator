/*
Copyright 2025 Wissam AA.

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

package controller

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/wissamir/appscaler-operator/api/v1alpha1"
)

// AppScalerReconciler reconciles a AppScaler object
type AppScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=api.operator.wissam.com,resources=appscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.operator.wissam.com,resources=appscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.operator.wissam.com,resources=appscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *AppScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling AppScaler")

	as := &apiv1alpha1.AppScaler{}
	err := r.Get(ctx, req.NamespacedName, as) // WA - what is this Get function?
	if err != nil {
		return ctrl.Result{}, nil
	}

	replicas := as.Spec.Replicas

	log.Info(fmt.Sprintf("Replicas %d", replicas))

	for _, deploy := range as.Spec.Deployments {
		d := &v1.Deployment{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: deploy.Namespace,
			Name:      deploy.Name,
		}, d)
		if err != nil {
			return ctrl.Result{}, err
		}
		if d.Spec.Replicas != &replicas {
			d.Spec.Replicas = &replicas
			err := r.Update(ctx, d)
			if err != nil {
				as.Status.Status = apiv1alpha1.FAILED
				return ctrl.Result{}, err
			}

			as.Status.Status = apiv1alpha1.SUCCESS
			err = r.Status().Update(ctx, as)
			if err != nil {
				return ctrl.Result{}, err
			}

		}
	}

	return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.AppScaler{}).
		Complete(r)
}
