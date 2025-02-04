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

package cri

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opencontainers/runtime-spec/specs-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubelet/pkg/types"
)

// Runtime represents a connection to the CRI-O runtime
type Runtime struct {
	Client cri.RuntimeServiceClient
}

// New returns a connection to the CRI runtime
func NewRuntime(socketPath string, timeout time.Duration) (*Runtime, error) {
	if socketPath == "" {
		return nil, fmt.Errorf("path to CRI socket missing")
	}

	clientConnection, err := connect(socketPath, timeout)
	if err != nil {
		return nil, fmt.Errorf("error establishing connection to CRI: %w", err)
	}

	return &Runtime{
		Client: cri.NewRuntimeServiceClient(clientConnection),
	}, nil
}

func (r *Runtime) NetworkNamespace(ctx context.Context, podName string, podNamespace string) (string, error) {
	podSandboxId, err := r.PodSandboxID(ctx, podName, podNamespace)
	if err != nil {
		return "", err
	}

	podSandboxStatus, err := r.Client.PodSandboxStatus(ctx, &cri.PodSandboxStatusRequest{
		PodSandboxId: podSandboxId,
		Verbose:      true,
	})
	if err != nil || podSandboxStatus == nil {
		return "", fmt.Errorf("failed to PodSandboxStatus for PodSandboxId %s: %w", podSandboxId, err)
	}

	sandboxInfo := &PodSandboxStatusInfo{}
	if err := json.Unmarshal([]byte(podSandboxStatus.Info[InfoKey]), sandboxInfo); err != nil {
		return "", fmt.Errorf("failed to Unmarshal podSandboxStatus.Info['%s']: %w", InfoKey, err)
	}

	networkNamespace := ""

	for _, namespace := range sandboxInfo.RuntimeSpec.Linux.Namespaces {
		if namespace.Type != specs.NetworkNamespace {
			continue
		}

		networkNamespace = namespace.Path
		break
	}

	if networkNamespace == "" {
		return "", fmt.Errorf("failed to find network namespace for PodSandboxId %s: %w", podSandboxId, err)
	}

	return networkNamespace, nil
}

func (r *Runtime) PodSandboxID(ctx context.Context, podName string, podNamespace string) (string, error) {
	// Labels used by Kubernetes: https://github.com/kubernetes/kubernetes/blob/v1.29.2/staging/src/k8s.io/kubelet/pkg/types/labels.go#L19
	ListPodSandboxRequest, err := r.Client.ListPodSandbox(ctx, &cri.ListPodSandboxRequest{
		Filter: &cri.PodSandboxFilter{
			LabelSelector: map[string]string{
				types.KubernetesPodNameLabel:      podName,
				types.KubernetesPodNamespaceLabel: podNamespace,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to ListPodSandbox for pod %s.%s: %w", podName, podNamespace, err)
	}

	if ListPodSandboxRequest == nil || ListPodSandboxRequest.Items == nil || len(ListPodSandboxRequest.Items) == 0 {
		return "", fmt.Errorf("ListPodSandbox returned 0 item for pod %s.%s: %w", podName, podNamespace, err)
	}

	if len(ListPodSandboxRequest.Items) > 1 {
		return "", fmt.Errorf("ListPodSandbox returned more than 1 item for pod %s.%s: %w", podName, podNamespace, err)
	}

	return ListPodSandboxRequest.Items[0].Id, nil
}

func connect(socketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	if socketPath == "" {
		return nil, fmt.Errorf("endpoint is not set")
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()

	conn, err := grpc.DialContext(
		ctx,
		criServerAddress(socketPath),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to endpoint '%s': %v", socketPath, err)
	}

	return conn, nil
}

func criServerAddress(criSocketPath string) string {
	return fmt.Sprintf("unix://%s", criSocketPath)
}
