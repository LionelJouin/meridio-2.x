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

package cli

import "github.com/spf13/cobra"

const (
	defaultVerbosity = 5
)

// CommonOptions represents all common options used in a run command CLI.
type CommonOptions struct {
	LogLevel int
}

// SetCommonFlags sets the flags for the common options used in a run command CLI.
func (co *CommonOptions) SetCommonFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(
		&co.LogLevel,
		"verbosity",
		"v",
		defaultVerbosity,
		"Log level, increase the increase the level increases the verbosity.")
}
