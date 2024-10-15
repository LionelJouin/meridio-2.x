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

package podnetworkannotator

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/pkg/controller/podnetworkannotator/network"
	"github.com/lioneljouin/meridio-experiment/pkg/endpointslice"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Controller reconcile the pods to add the network resources required
// by the service (network attachement, VIP, routes...).
type Controller struct {
	client.Client
	Scheme           *runtime.Scheme
	GatewayClassName string
	// GetIPsFunc is used when the endpointSlice will be reconciled to get the IPs
	// of the pods attached to the service.
	GetIPsFunc       endpointslice.GetIPs
	NetworkAnnotator *network.Annotater
}

// Reconcile implements the reconciliation of the pod object.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	pod := &v1.Pod{}

	err := c.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get the pod: %w", err)
	}

	if pod.Status.Phase != v1.PodRunning {
		return ctrl.Result{}, nil
	}

	err = c.reconcileGateways(ctx, pod)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile the gateways: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1.Pod{}).
		// With EnqueueRequestsFromMapFunc, on an update the func is called twice
		// (1 time for old and 1 time for new object)
		Watches(&v1.Service{}, handler.EnqueueRequestsFromMapFunc(c.serviceEnqueue)).
		Watches(&gatewayapiv1.Gateway{}, handler.EnqueueRequestsFromMapFunc(c.gatewayEnqueue)).
		Watches(&v1.Pod{}, handler.EnqueueRequestsFromMapFunc(c.podEnqueue)).
		Complete(c)
	if err != nil {
		return fmt.Errorf("failed to build the pod controller manager: %w", err)
	}

	return nil
}
