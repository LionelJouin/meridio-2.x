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

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/cli"
	"github.com/lioneljouin/meridio-experiment/pkg/controller/gateway"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type runOptions struct {
	cli.CommonOptions
	gatewayClassName string
}

func newCmdRun() *cobra.Command {
	runOpts := &runOptions{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the controller-manager",
		Long:  `Run the controller-manager`,
		Run: func(cmd *cobra.Command, _ []string) {
			runOpts.run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(
		&runOpts.gatewayClassName,
		"gateway-class-name",
		"",
		"Name of the Gateway Class handled by this controller manager.",
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

	logger := log.New("controller-manager", ro.LogLevel)

	crlog.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false,
		Cache:          cache.Options{},
		Metrics: server.Options{
			BindAddress: "0",
		},
		HealthProbeBindAddress: ":8081",
	})
	if err != nil {
		log.Fatal(setupLog, "failed to create manager for controllers", "err", err)
	}

	if err = (&gateway.Controller{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		GatewayClassName: ro.gatewayClassName,
	}).SetupWithManager(mgr); err != nil {
		log.Fatal(setupLog, "failed to create controller", "err", err, "controller", "Gateway")
	}

	// if err = (&podinjector.Controller{
	// 	Client:           mgr.GetClient(),
	// 	Scheme:           mgr.GetScheme(),
	// 	GatewayClassName: ro.gatewayClassName,
	// 	GetIPsFunc:       networkattachment.GetIPs,
	// }).SetupWithManager(mgr); err != nil {
	// 	log.Fatal(setupLog, "failed to create controller", "err", err, "controller", "Pod")
	// }

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
