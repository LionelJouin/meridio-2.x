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

// Package cmd provides the CLI for the router program.
package cmd

import (
	"fmt"
	"os"

	"github.com/lioneljouin/meridio-experiment/pkg/cli"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by the main function.
func Execute() {
	if err := getRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "router",
		Short: "CLI",
		Long:  `CLI for interacting with the router`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	rootCmd.AddCommand(newCmdRun())
	rootCmd.AddCommand(cli.NewCmdVersion())

	return rootCmd
}
