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

package bird

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var errBirdRunning = errors.New("bird is already running")

// Bird represents the bird configuration.
type Bird struct {
	// SocketPath is the full path with filename to communicate with birdc
	SocketPath string
	// configuration file (with path)
	ConfigFile string
	// Add important bird log snippets to bird logs
	LogEnabled bool
	// Size of the file for the bird logs
	LogFileSize   int
	LogFile       string
	LogFileBackup string

	running bool
	mu      sync.Mutex
}

// New is the bird constructor.
func New() *Bird {
	return &Bird{
		SocketPath:    "/var/run/bird/bird.ctl",
		ConfigFile:    "/etc/bird/bird.conf",
		LogEnabled:    true,
		LogFileSize:   defaultLogFileSize,
		LogFile:       "/var/log/bird.log",
		LogFileBackup: "/var/log/bird.log.backup",
	}
}

// Run starts bird with the current bird configuration. Bird will be stopped
// when the context in parameter will be cancelled.
func (b *Bird) Run(ctx context.Context) error {
	b.mu.Lock()

	if b.running {
		return errBirdRunning
	}

	// Write empty config if config file does not exist
	if _, err := os.Stat(b.ConfigFile); errors.Is(err, os.ErrNotExist) {
		err := b.writeConfig([]string{}, []Gateway{})
		if err != nil {
			return err
		}
	}

	b.running = true

	b.mu.Unlock()

	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"bird",
		"-d",
		"-c",
		b.ConfigFile,
		"-s",
		b.SocketPath,
	)

	var errFinal error

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil && !errors.Is(err, context.Cause(ctx)) {
		errFinal = fmt.Errorf("failed starting bird ; %w; %s", err, stdoutStderr)
	}

	return errFinal
}

// Configure writes the bird configuration file, sets the policy routes and configures bird if it is running.
func (b *Bird) Configure(ctx context.Context, vips []string, gateways []Gateway) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	err := b.writeConfig(vips, gateways)
	if err != nil {
		return err
	}

	if b.running {
		//nolint:gosec
		cmd := exec.CommandContext(
			ctx,
			"birdc",
			"-s",
			b.SocketPath,
			"configure",
			`"`+b.ConfigFile+`"`,
		)

		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil && !errors.Is(err, context.Cause(ctx)) {
			return fmt.Errorf("failed configuring bird ; %w; %s", err, stdoutStderr)
		}
	}

	err = setPolicyRoutes(vips)
	if err != nil {
		return fmt.Errorf("failed to set the policy routes %v: %w", vips, err)
	}

	return nil
}

func (b *Bird) writeConfig(vips []string, gateways []Gateway) error {
	file, err := os.Create(b.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to create %v, err: %w", b.ConfigFile, err)
	}

	defer file.Close()

	conf := b.getConfig(vips, gateways)

	_, err = file.WriteString(conf)
	if err != nil {
		return fmt.Errorf("failed to write config to %v, err: %w", b.ConfigFile, err)
	}

	return nil
}
