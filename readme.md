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

Install Whereabouts:
```
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/whereabouts/refs/tags/v0.8.0/doc/crds/daemonset-install.yaml
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/whereabouts/refs/tags/v0.8.0/doc/crds/whereabouts.cni.cncf.io_ippools.yaml
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/whereabouts/refs/tags/v0.8.0/doc/crds/whereabouts.cni.cncf.io_overlappingrangeipreservations.yaml
```

Install Meridio Experiment:
```
helm install poc ./deployments/PoC --set registry=ghcr.io/lioneljouin/meridio-experiment --set imagePullPolicy=IfNotPresent --set sllbReplicas=2
```

Install Gateways/Routers/Traffic-Generators (`docker compose down` to uninstall. Change manually the image in `docker-compose.yaml` if you built your own):
```
docker compose up -d
```

## Demo

Install the example Gateway/GatewayRouter/Service:
```
kubectl apply -f examples/sllb-a.yaml
kubectl apply -f examples/sllb-b.yaml
```

Install example application behind the service:
```
helm install example-target-application-multi ./examples/target-application/deployment/helm --set applicationName=multi --set networks='[{"name":"macvlan-nad-1","interface":"net1"},{"name":"macvlan-nad-2","interface":"net2"}]' --set registry=ghcr.io/lioneljouin/meridio-experiment
```

```
docker exec -it vpn-a mconnect -address 20.0.0.1:4000 -nconn 400 -timeout 2s
docker exec -it vpn-b mconnect -address 40.0.0.1:4000 -nconn 400 -timeout 2s
```