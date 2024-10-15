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

package enqueuer

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ServiceEnqueue(
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
