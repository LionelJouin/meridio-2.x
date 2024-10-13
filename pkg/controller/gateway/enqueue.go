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

package gateway

import (
	"context"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func serviceEnqueue(
	_ context.Context,
	object client.Object,
) []reconcile.Request {
	gatewayName, exists := object.GetLabels()[apis.LabelServiceProxyName]
	if !exists {
		return []reconcile.Request{}
	}

	// todo: check if parent is the right class

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      gatewayName,
				Namespace: object.GetNamespace(),
			},
		},
	}
}

func l34RouteEnqueue(
	_ context.Context,
	object client.Object,
) []reconcile.Request {
	l34Route, ok := object.(*v1alpha1.L34Route)
	if !ok {
		return []reconcile.Request{}
	}

	if len(l34Route.Spec.ParentRefs) == 0 {
		return []reconcile.Request{}
	}

	// todo: check if parent is the right class

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      string(l34Route.Spec.ParentRefs[0].Name),
				Namespace: object.GetNamespace(),
			},
		},
	}
}
