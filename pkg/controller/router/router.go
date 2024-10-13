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

package router

import (
	"context"
	"fmt"
	"net"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/bird"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/proxy/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// RoutingSuite defines an interface to configure a routing suite (e.g. bird).
type RoutingSuite interface {
	// Apply applies to configuration.
	Configure(ctx context.Context, vips []string, gateways []bird.Gateway) error
}

// Controller reconciles the Gateway Object to run a router.
type Controller struct {
	client.Client
	Scheme *runtime.Scheme
	// Name of the gateway in which this controller is running.
	Name string
	// Namespace of the gateway in which this controller is running.
	Namespace            string
	RoutingSuiteInstance RoutingSuite
}

// Reconcile implements the reconciliation of the Gateway of the router.
// This function is trigger by any change (create/update/delete) in any resource related
// to the object (GatewayRouter/Gateway).
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	gateway := &gatewayapiv1.Gateway{}

	if req.Name != c.Name || req.Namespace != c.Namespace {
		// this should not happen if the controller is configured correctly.
		return ctrl.Result{}, nil
	}

	err := c.Get(ctx, req.NamespacedName, gateway)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get the gateway: %w", err)
	}

	vips := getVIPs(gateway)

	gateways, err := c.getGatewayRouters(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get the gateway routers: %w", err)
	}

	log.FromContextOrGlobal(ctx).Info("Configure RoutingSuiteInstance", "vips", vips, "gateways", gateways)

	err = c.RoutingSuiteInstance.Configure(ctx, vips, gateways)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set the gateway: %w", err)
	}

	return ctrl.Result{}, nil
}

// getVIPs gets the list of addresses in the gateway status and converts them to CIDRs.
func getVIPs(gateway *gatewayapiv1.Gateway) []string {
	vips := []string{}

	// remove duplicates and invalid VIPs
	vipMap := map[string]struct{}{}

	for _, address := range gateway.Status.Addresses {
		ipAddr := net.ParseIP(address.Value)

		if ipAddr == nil {
			continue
		}

		vip := fmt.Sprintf("%s/32", ipAddr)

		if ipAddr.To4() == nil {
			vip = fmt.Sprintf("%s/128", ipAddr)
		}

		_, exists := vipMap[vip]
		if exists {
			continue
		}

		vips = append(vips, vip)

		vipMap[vip] = struct{}{}
	}

	return vips
}

// getGatewayRouters gets the list of gateways for this gateway.
func (c *Controller) getGatewayRouters(ctx context.Context) ([]bird.Gateway, error) {
	gatewayList := &v1alpha1.GatewayRouterList{}
	gateways := []bird.Gateway{}

	err := c.List(ctx,
		gatewayList,
		client.MatchingLabels{
			apis.LabelServiceProxyName: c.Name,
		}) // todo: filter namespace
	if err != nil {
		return nil, fmt.Errorf("failed listing the flows: %w", err)
	}

	for _, gateway := range gatewayList.Items {
		g := gateway
		gateways = append(gateways, newGateway(&g))
	}

	return gateways, nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapiv1.Gateway{}).
		// With EnqueueRequestsFromMapFunc, on an update the func is called twice
		// (1 time for old and 1 time for new object)
		Watches(&v1alpha1.GatewayRouter{}, handler.EnqueueRequestsFromMapFunc(gatewayRouterEnqueue)).
		Complete(c)
	if err != nil {
		return fmt.Errorf("failed to build the Router manager: %w", err)
	}

	return nil
}
