# Meridio Experiment

### Build

This builds, tags and pushes
```
make generate
make REGISTRY=ghcr.io/lioneljouin/meridio-experiment
```

### pre-requisites:

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

```
helm install poc ./deployments/PoC --set registry=ghcr.io/lioneljouin/meridio-experiment
```
