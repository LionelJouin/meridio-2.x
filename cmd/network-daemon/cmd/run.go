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

package cmd

import (
	"context"
	"time"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/cli"
	"github.com/lioneljouin/meridio-experiment/pkg/controller/podnetworkinjector"
	"github.com/lioneljouin/meridio-experiment/pkg/cri"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type runOptions struct {
	cli.CommonOptions
	namespace     string
	criSocketPath string
	nodeName      string
}

func newCmdRun() *cobra.Command {
	runOpts := &runOptions{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the network-daemon",
		Long:  `Run the network-daemon`,
		Run: func(cmd *cobra.Command, _ []string) {
			runOpts.run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(
		&runOpts.namespace,
		"namespace",
		"default",
		"namespace of the gateway in which the network-daemon is running.",
	)

	cmd.Flags().StringVar(
		&runOpts.criSocketPath,
		"cri-socket-path",
		"/run/containerd/containerd.sock",
		"Path to the CRI socket.",
	)

	cmd.Flags().StringVar(
		&runOpts.nodeName,
		"node-name",
		"",
		"Name of the node where this pod is running.",
	)

	runOpts.SetCommonFlags(cmd)

	return cmd
}

func (ro *runOptions) run(ctx context.Context) {
	scheme := runtime.NewScheme()
	setupLog := ctrl.Log.WithName("setup")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(gatewayapiv1.Install(scheme))

	logger := log.New("network-daemon", ro.LogLevel)

	crlog.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&v1.Pod{}: {
					// Field: fields.OneTermEqualSelector("spec.nodeName", ro.nodeName),
					// Field: fields.OneTermEqualSelector("spec.hostNetwork", ),
					Field: fields.SelectorFromSet(fields.Set{
						"spec.nodeName":    ro.nodeName,
						"spec.hostNetwork": "false",
					}),
				},
			},
		},
		Metrics: server.Options{
			BindAddress: "0",
		},
		HealthProbeBindAddress: ":8082",
	})
	if err != nil {
		log.Fatal(setupLog, "failed to create manager for controllers", "err", err)
	}

	const shortTimeout = 5 * time.Second
	containerRuntime, err := cri.NewRuntime(ro.criSocketPath, shortTimeout)
	if err != nil {
		log.Fatal(setupLog, "failed to connect to container runtime", "err", err)
	}

	if err = (&podnetworkinjector.Controller{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		MinTableID:       50000,
		MaxTableID:       55000,
		ContainerRuntime: containerRuntime,
	}).SetupWithManager(mgr); err != nil {
		log.Fatal(setupLog, "failed to create controller", "err", err, "controller", "Gateway")
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal(setupLog, "unable to set up health check", "err", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal(setupLog, "unable to set up ready check", "err", err)
	}

	if err := mgr.Start(ctx); err != nil {
		log.Fatal(setupLog, "failed to start manager", "err", err)
	}
}
