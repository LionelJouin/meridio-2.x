/*
Copyright (c) 2024 OpenInfra Foundation Europe

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

package endpointslice

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/pkg/controller/endpointslice/enqueuer"
	"github.com/lioneljouin/meridio-experiment/pkg/controller/endpointslice/reconciler"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Controller reconciles the Service Object.
type Controller struct {
	client.Client
	Scheme            *runtime.Scheme
	GatewayClassName  string
	ServiceReconciler *reconciler.Service
}

// Reconcile implements the reconciliation of the Service to create the associated EndpointSlice.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	gateway := &gatewayapiv1.Gateway{}

	err := c.Get(ctx, req.NamespacedName, gateway)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get the gateway: %w", err)
	}

	if string(gateway.Spec.GatewayClassName) != c.GatewayClassName {
		// this should not happen if the controller is configured correctly.
		return ctrl.Result{}, nil
	}

	err = c.ServiceReconciler.Reconcile(ctx, gateway)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile the services: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	enqr := enqueuer.New(c.Client, c.GatewayClassName)

	err := ctrl.NewControllerManagedBy(mgr).
		Named("EndpointSlice").
		For(&gatewayapiv1.Gateway{}).
		Owns(&v1discovery.EndpointSlice{}).
		Watches(&v1.Service{}, handler.EnqueueRequestsFromMapFunc(enqueuer.ServiceEnqueue)).
		Watches(&v1.Pod{}, handler.EnqueueRequestsFromMapFunc(enqr.PodEnqueue)).
		Complete(c)
	if err != nil {
		return fmt.Errorf("failed to build the Service EndpointSlice manager: %w", err)
	}

	return nil
}
