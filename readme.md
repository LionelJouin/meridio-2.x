# Meridio Experiment

### Build

This builds, tags and pushes
```
make generate
make REGISTRY=ghcr.io/lioneljouin/meridio-experiment
```

### Pre-Requisites

Install Gateway API:
```
kubectl apply -k https://github.com/kubernetes-sigs/gateway-api/config/crd/experimental?ref=v1.1.0
```

Install CNI Plugins:
```
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/e2e/templates/cni-install.yml.j2
```

Install Multus:
```
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/refs/heads/master/deployments/multus-daemonset.yml
```

Install Meridio Experiment:
```
helm install poc ./deployments/PoC --set registry=ghcr.io/lioneljouin/meridio-experiment --set imagePullPolicy=IfNotPresent --set sllbReplicas=2
```

## Demo

Install the example Gateway/GatewayRouter/Service:
```
kubectl apply -f examples/sllb-a.yaml
```

Install example application behind the service:
```
helm install example-target-application-a ./examples/target-application/deployment/helm --set applicationName=a --set registry=ghcr.io/lioneljouin/meridio-experiment
```