# Meridio 2.x

Meridio 2.x is an evolution of [Nordix/Meridio](https://github.com/nordix/meridio) built upon its foundation to offer a more modern and efficient approach. It enhances existing capabilities while introducing new features to better align with current technological trends and requirements.

As of today (January 2025), this project is in the Proof of Concept (PoC) phase with ongoing development and testing to validate its viability and performance before moving into a broader implementation.

An older implementation of this PoC is existing here: [LionelJouin/l-3-4-gateway-api-poc](https://github.com/LionelJouin/l-3-4-gateway-api-poc).

## Table of Contents

<!-- Generated with https://bitdowntoc.derlin.ch/ -->
* [Summary](#summary)
* [Motivation](#motivation)
    + [Why Meridio](#why-meridio)
* [API](#api)
    + [Gateway API and Kubernetes API](#gateway-api-and-kubernetes-api)
        - [Gateway](#gateway)
            * [GatewayClass](#gatewayclass)
        - [L34Route](#l34route)
        - [Service](#service)
    + [GatewayRouter](#gatewayrouter)
    + [API Nordix/Meridio Differences](#api-nordixmeridio-differences)
* [Components](#components)
    + [Stateless-Load-Balancer-Router (SLLBR)](#stateless-load-balancer-router-sllbr)
        - [Stateless-Load-Balancer (SLLB)](#stateless-load-balancer-sllb)
        - [Router](#router)
    + [Controller-Manager](#controller-manager)
    + [Alternatives for Application Network Configuration Injection](#application-network-configuration-injection)
        - [Annotation Injection](#annotation-injection)
            * [Dynamic Network Attachment](#dynamic-network-attachment)
            * [Network Daemon](#network-daemon)
        - [Sidecar](#sidecar)
        - [Static Configuration](#static-configuration)
    + [Components Nordix/Meridio Differences](#components-nordixmeridio-differences)
* [Data Plane](#data-plane)
    + [DC Gateway to Meridio Gateway](#dc-gateway-to-meridio-gateway)
    + [Stateless-Load-Balancer-Router (SLLBR)](#stateless-load-balancer-router-sllbr-1)
    + [Meridio Gateway to Endpoint (Application pod)](#meridio-gateway-to-endpoint-application-pod)
    + [Endpoint (Application pod)](#endpoint-application-pod)
    + [Dataplane Nordix/Meridio Differences](#dataplane-nordixmeridio-differences)
* [Extra Features](#extra-features)
    + [Port Address Translation (PAT)](#port-address-translation-pat)
    + [Gateway Scaling](#gateway-scaling)
* [Prerequisites](#prerequisites)
* [Multi-Tenancy](#multi-tenancy)
* [Upgrade and Migration](#upgrade-and-migration)
    + [Data Plane](#data-plane-1)
    + [Meridio 2.x](#meridio-2x-1)
    + [From Nordix/Meridio](#from-nordixmeridio)
* [Project Structure and Implementation](#project-structure-and-implementation)
    + [Projects](#projects)
    + [Framework amd Design Pattern](#framework-amd-design-pattern)
* [Evolution](#evolution)
    + [Dynamic Resource Allocation](#dynamic-resource-allocation)
    + [Service Type](#service-type)
    + [Service Chaining](#service-chaining)
* [Alternatives](#alternatives)
    + [LoxiLB](#loxilb)
    + [OVN-Kubernetes](#ovn-kubernetes)
    + [F5](#f5)
    + [Google](#google)
    + [Cilium](#cilium)

## Summary

This document provides an overview of the architecture of the project focusing on its core components and how they work together to deliver their functionalities. It outlines the relationships between different elements within the system offering an understanding of their roles and interactions. This document also serves as a practical guide for deployment detailing the required infrastructure, dependencies and configuration steps needed to get the system up and running.

Beyond deployment, this document explores potential areas for growth and evolution considering how the architecture can adapt to new use cases, include new features, and integrate with emerging technologies. By anticipating future challenges and opportunities, the project aims to remain relevant and adaptable in an ever-changing cloud-native technological landscape.

Finally, this document provides an analysis of existing alternatives examining their strengths and weaknesses in comparison to the proposed solution. This comparison helps to highlight the unique advantages of the project and offers insights into how it differentiates itself from other approaches in the ecosystem.

## Motivation

The motivation behind this project is to evolve [Nordix/Meridio](https://github.com/nordix/meridio) and improve the overall system by utilizing industry standards such as, for example, Multus as the main secondary network provider. The design and implementation is streamlined focusing on reducing complexity while maintaining the core functionalities.

This evolution aims to simplify every aspect of the project starting with the user configuration by providing a more intuitive API, therefore making the system easier to use and configure. 

Key components are being pushed upstream to leverage community-driven improvements, such as [Kubernetes](https://github.com/kubernetes), the [Kubernetes Network Plumbing Working Group](https://github.com/k8snetworkplumbingwg), and [CNI](https://github.com/containernetworking). This reduces the need for custom solutions and ensures greater alignment with broader industry trends. 

Unnecessary components such as [Proxy](https://meridio.nordix.org/docs/components/proxy), [IPAM](https://meridio.nordix.org/docs/components/ipam) and [TAPA](https://meridio.nordix.org/docs/components/tapa) that were previously under the responsibilities of [Nordix/Meridio](https://github.com/nordix/meridio) are removed allowing a more focused and efficient system. Reducing the number of components within the system also minimizes potential points of failure and simplifies maintenance and troubleshooting.

Deprecated and unsuitable components, such as Network Service Mesh, are phased out to make a way towards more modern solutions. This will be causing unused dependencies to be also eliminated (e.g. Spire and Open Telemetry), freeing the project from unnecessary bloat.

The evolution also enables an enhanced networking control giving users more flexibility and customization options, thus allowing the project to be deployed on a wider range of PaaS platforms, increasing its versatility and accessibility.

### Why Meridio

TODO

## API

### Gateway API and Kubernetes API

The project aims to leverage standard Kubernetes APIs as much as possible and focuses on providing implementations that seamlessly integrate with existing workflows. By adopting [Gateway API](https://github.com/kubernetes-sigs/gateway-api), the project aligns with Kubernetes-native networking paradigms, ensuring compatibility with upstream developments and reducing the need for custom solutions.

#### Gateway

The `Gateway` object represents an instance of a service-traffic handling infrastructure. This API object has been officially released as version 1 (v1). More can be read about it on the official documentation: [gateway-api.sigs.k8s.io/concepts/api-overview/#gateway](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway).

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: sllb-a
spec:
  gatewayClassName: meridio-experiment/stateless-load-balancer
  listeners: # This is to be ignored
  - name: all 
    port: 4000
    protocol: TCP
  infrastructure:
    annotations: 
      k8s.v1.cni.cncf.io/networks: '[{"name":"vlan-100","interface":"vlan-100"},{"name":"macvlan-nad-1","interface":"net1"}]' # Networks attached to the gateway workloads
      meridio-experiment/networks: '[{"name":"macvlan-nad-1","interface":"net1"}]' # Networks towards the service selected pods
      meridio-experiment/network-subnets: '["169.111.100.0/24"]'
```

Here is an example above of how a `Gateway` would look like. Once deployed, the operator/controller will be in charge of deploying the component for the gateway to provide the functionalities described by the GatewayClass `meridio-experiment/stateless-load-balancer`. 

The `.spec.infrastructure` field represents some configuration specific to the workload being deployed. The `k8s.v1.cni.cncf.io/networks` will be added to the `Gateway` workloads indicating the pod will be attached to these networks via Multus. `meridio-experiment/networks` and `meridio-experiment/network-subnets` will indicate that the IPs of the endpoints behind the services being handled by the `Gateway` must be in the subnet `169.111.100.0/24` on the network provided by Multus via `macvlan-nad-1`.

For more control, if required in the future, another sub-field named `parametersRef` allows passing a configuration object.

The `.spec.listeners` field is not relevant for the functionalities provided by this project and must be ignored. Gateway API enforces to have at least one item, so a random one must be set. An option which could make it relevant to this project has mentioned it these conversations: [kubernetes-sigs/gateway-api#130](https://github.com/kubernetes-sigs/gateway-api/pull/130), [kubernetes-sigs/gateway-api#780](https://github.com/kubernetes-sigs/gateway-api/pull/780#discussion_r693210917), [kubernetes-sigs/gateway-api#818](https://github.com/kubernetes-sigs/gateway-api/issues/818) and [kubernetes-sigs/gateway-api#1061](https://github.com/kubernetes-sigs/gateway-api/issues/1061).

Once deployed, the `.status.Conditions` field will be set indicating the status of the `Gateway` (Ready if deployed successfully, and NotReady if not).

The keys of the annotations in the `.spec.infrastructure` field are likely to change, but the concepts and functionalities will remain the same.

##### GatewayClass

The `GatewayClass` represents the type of `Gateway` which can be deployed in the cluster. The existence of this object is not enforced by Gateway API. More can be read about it on the official documentation: [gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass).

```yaml
---
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: meridio-experiment/stateless-load-balancer
spec:
  controllerName: meridio-experiment/stateless-load-balancer
  description: "Telco stateless Load-Balancer handling L34Routes and Services."
```

Here is an example of how this `GatewayClass` could look like. In the PoCs, this object is not used.

#### L34Route

The `L34Route` provides a mechanism to route Layer 3 and Layer 4 traffic to a designated backend, typically represented as a Kubernetes `Service`. It enables precise traffic steering based on IP address and transport-layer attributes ensuring that packets matching specified criteria are forwarded to the appropriate destination.

The `.spec.parentRefs` field within an L34Route resource specifies the `Gateway` in which the route will be configured. Once the route is established within a `Gateway`, the `.spec.destinationCIDRs` defined in the configuration are reflected in the `.status.Addresses` field of the `Gateway`. This status update indicates that the `Gateway` is actively handling traffic for the specified addresses.

To maintain a clear routing setup, `L34Route` enforces certain constraints: each route can reference only a single `.spec.parentRefs` and a single `.spec.backendRefs` ensuring a straightforward and predictable traffic flow. Additionally, only `/32` CIDRs for IPv4 and `/128` CIDRs for IPv6 are supported in the `.spec.destinationCIDRs` field.

Additionally, multiple `L34Route` resources can be configured within the same `Gateway` and multiple `L34Route` can reference the same `.spec.backendRefs`. In such cases, the `L34Route` with the highest `.spec.priority` value is prioritized.

```yaml
apiVersion: meridio.experiment.gateway.api.poc/v1alpha1
kind: L34Route
metadata:
  name: vip-20-0-0-1-multi-ports-a
spec:
  parentRefs:
  - name: sllb-a
  backendRefs:
  - name: service-a
    port: 1 # This must be set but is being ignored.
  priority: 10
  destinationCIDRs:
  - 20.0.0.1/32
  sourceCIDRs:
  - 0.0.0.0/0
  sourcePorts:
  - 0-65535
  destinationPorts:
  - "4000"
  - "4001"
  protocols:
  - TCP
```

In the above example, the L34Route is configured in the `Gateway` called `sllb-a`. The TCP IPv4 traffic with any source IP (`0.0.0.0/0`), any source port (port range `0-65535`), a destination IP `20.0.0.1` (`20.0.0.1/32`) and a port `4000` or `4001` will be steered towards the backend `service-a`.

This API is not part of Gateway API but discussions are happening within the community about it here: [User reports for TCPRoute and UDPRoute (kubernetes-sigs/gateway-api/discussions#3475)](https://github.com/kubernetes-sigs/gateway-api/discussions/3475#discussioncomment-11522199) and [Layer 3 / load balancer / Service route (kubernetes-sigs/gateway-api/discussions#3351)](https://github.com/kubernetes-sigs/gateway-api/discussions/3351#discussioncomment-11522379).

The names of the attributes are likely to change in this API, but the concepts and functionalities will remain the same.

#### Service

In the context of this project, the `Service` object from the Kubernetes Core API is used as a headless Service, which means the Service is configured without a cluster IP. This configuration allows the system to bypass the default Kubernetes load balancing and instead rely on a custom load-balancer.

The `Service` specifies in which `Gateway` it must be running via the well-known `Service` label `service.kubernetes.io/service-proxy-name`, thus preventing a `Service` to be running in several `Gateway`. The endpoints of the `Service` are selected via the `.spec.selector` field.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: service-a
  labels:
    service.kubernetes.io/service-proxy-name: sllb-a
spec:
  clusterIP: None
  selector:
    app: example-target-application-multi
    meridio-experiment/dummy-service-selector: "true"
```

In the example above, the `Service` is set to be configured by and running in the `Gateway` named `sllb-a`. The `Service` will load-balance the traffic it will receive to the endpoints (pods) with the label `app=example-target-application-multi`. 

The label `meridio-experiment/dummy-service-selector=true` is ignored by the implementation. The reason to have the label is to disable the main Kubernetes EndpointSlice Controller which selects the primary pod IPs. The [KEP 4770 (EndpointSlice Controller Flexibility)](https://github.com/kubernetes/enhancements/issues/4770) is proposing a solution to avoid this situation and provide a native Kubernetes feature.

Discussions are currently happening in the Gateway API community towards an alternative to `Service` as a backend. The discussion ([Google Doc](https://docs.google.com/document/d/1N-C-dBHfyfwkKufknwKTDLAw4AP2BnJlnmx0dB-cC4U/edit?tab=t.0#heading=h.7a7xx7cbdmx2)) is about a new object named `EndpointSelector` which would provide a similar `selector` field as the `Service` with more possibility to extend it for modern solutions.

### GatewayRouter

The `GatewayRouter` defines how to establish connectivity with an external DC-Gateway such as, which IP-address, routing- and supervision-protocol(s) to use. The `GatewayRouter` can also define specific protocol settings that differ from the default.

The `Gateway` in which the `GatewayRouter` is being configured in is specified via the label `service.kubernetes.io/service-proxy-name`.

Except the `.spec.interface` and the label, this `GatewayRouter` object is similar to the `Gateway` object in Nordix/Meridio.

```yaml
apiVersion: meridio.experiment.gateway.api.poc/v1alpha1
kind: GatewayRouter
metadata:
  name: gateway-a-v4
  labels:
    service.kubernetes.io/service-proxy-name: sllb-a
spec:
  address: 169.254.100.150
  interface: vlan-100
  bgp:
    localASN: 8103
    remoteASN: 4248829953
    holdTime: 24s
    localPort: 10179
    remotePort: 10179
    bfd:
      switch: true
      minTx: 300ms
      minRx: 300ms
      multiplier: 5
```

### API Nordix/Meridio Differences

Here below, a table summarizing the API differences:

| Nordix/Meridio | Meridio 2.x | Description |
| -------- | ------- | ------- |
| [Trench](https://meridio.nordix.org/docs/concepts/trench) | N/A |  |
| [VIP](https://meridio.nordix.org/docs/concepts/vip) | N/A |  |
| [Gateway](https://meridio.nordix.org/docs/concepts/gateway) | GatewayRouter | Connects to the external Router/DC-Gateway and advertises VIPs |
| [Attractor](https://meridio.nordix.org/docs/concepts/attractor) | Gateway | Workload handling traffic where the `L34Routes` and `Services` are running |
| [Conduit](https://meridio.nordix.org/docs/concepts/conduit) | N/A |  |
| [Stream](https://meridio.nordix.org/docs/concepts/stream) | Service | Load-Balances traffic among endpoints (application pods) |
| [Flow](https://meridio.nordix.org/docs/concepts/flow) | L34Route | Classifies traffic to a `Service` |

Here below, a traffic flow diagram of Nordix/Meridio:
![Overview](resources/diagrams-Flow-v1.png)

Here below, a traffic flow diagram of Meridio 2.x:
![Overview](resources/diagrams-Flow-v2.png)

## Components

### Stateless-Load-Balancer-Router (SLLBR)

The Stateless-Load-Balancer-Router (SLLBR) is the workload (`Deployment`) instance(s) behind a `Gateway` object handling service-traffic. It is designed to provide high-performance packet forwarding without maintaining connection state, ensuring scalability and resilience.

SLLBR is composed of two containers:
* `Stateless-Load-Balancer`: Responsible for distributing incoming traffic across backend pods based on defined `L34Routes` and `Services`.
* `Router`: Responsible for advertising VIPs to `GatewayRouter`.

To function correctly, SLLBR requires specific system settings (SYSCTLs) to be configured within the pod's network namespace. These settings include:
* `forwarding` set to `1` to enable IP forwarding.
* `fib_multipath_hash_policy` set to `1` to allow multipath routing based on layer 4 hash.
* `rp_filter` set to `0` to allow packets to have a source address which does not correspond to any routing destination address.
* `fwmark_reflect` set to `1` to allow generated outbound ICMP `fragmentation needed` reply to use VIP as source address.
* `ip_local_port_range` set to `49152 65535` to define the range of ephemeral ports available for outgoing connections to fulfill the BFD Control packet requirements.

#### Stateless-Load-Balancer (SLLB)

The Stateless-Load-Balancer container is responsible for running the load-balancer process based on [Nordix/nfqueue-loadbalancer](https://github.com/Nordix/nfqueue-loadbalancer) (NFQLB). This component efficiently distributes incoming traffic to backend pods without maintaining session state, ensuring high performance and scalability. 

The container continuously watches relevant Kubernetes objects such as `Gateway`, `L34Route`, `Service` and `EndpointSlices` to dynamically configure NFQLB, nftables and routes ensuring that changes in the cluster are reflected in the load-balancer configuration in real time. 

The readiness of the container is determined by two key factors: the successful execution of the NFQLB process and the ability to communicate with the Kubernetes API. This ensures that the load-balancer is fully operational and integrated within the cluster.

To function properly, the container requires specific Linux capabilities and Kubernetes API access, including:
* `NET_ADMIN` to configure network settings such as routing and firewall rules.
* `IPC_LOCK` to allow NFQLB to use shared memory.
* `IPC_OWNER` to allow NFQLB to use shared memory.
* `Kubernetes API`: `watch`, `list` and `get` the `Gateway`, `L34Route`, `Service` and `EndpointSlices` objects.

#### Router

The Router container is responsible for running the routing suite Bird2 and advertising the VIPs handled by the `Gateway` to the DC-Gateway.

The container continuously watches relevant Kubernetes objects such as `Gateway` and `GatewayRouter` to ensure that changes in the cluster are reflected in the Bird2 configuration in real time. 

The readiness of the container is determined by its ability to successfully run Bird2 and establish communication with the Kubernetes API.

To function properly, the container requires specific Linux capabilities and Kubernetes API access, including:
* `NET_ADMIN` to manage and modify routing tables.
* `NET_BIND_SERVICE` to allow Bird2 to bind to privileged ports.
* `NET_RAW` to Bird2 BIRD to use the SO_BINDTODEVICE socket option.
* `Kubernetes API`: `watch`, `list` and `get` the `Gateway` and `GatewayRouter` objects.

### Controller-Manager

The Controller-Manager (Operator) is responsible for reconciling and managing Kubernetes objects that belong to the appropriate `GatewayClass`, ensuring that the correct resources are deployed and maintained according to the specified configuration.

A key function of the Controller Manager is to reconcile and manage `Gateway` objects to deploy and maintain the Stateless-Load-Balancer-Router (SLLBR) component.

Additionally, the Controller Manager handles the reconciliation of `EndpointSlice` objects which consist in maintaining endpoint information (IP and Status) for the pods selected by the `Services` managed by the associated `Gateways`. This ensures that traffic is correctly routed to the appropriate backend instances maintaining service availability and performance.

The [KEP 4770 (EndpointSlice Controller Flexibility)](https://github.com/kubernetes/enhancements/issues/4770) was proposing a solution to transform the Kubernetes EndpointSlice Controller as a generic one capable of selecting secondary IPs. Since it got rejected, the alternative ([sig-network meeting notes from 2025-01-16](https://docs.google.com/document/d/1_w77-zG_Xj0zYvEMfQZTQ-wPP4kXkpGD8smVtW_qqWM/edit?tab=t.0)) is to implement it as a separated repository under `kubernetes-sigs` so it can be re-used for this use-case.

The readiness of the container is determined by its ability to successfully establish communication with the Kubernetes API.

To perform these operations, the Controller Manager requires access to a set of Kubernetes API resources:
* `Pod`: `watch`, `list` and `get`.
* `Deployment`: `create`, `delete`, `get`, `list`, `patch`, `update` and `watch`.
* `Gateway`: `watch`, `list` and `get`.
* `Gateway/status`: `patch` and `update`.
* `Service`: `watch`, `list` and `get`.
* `EndpointSlice`: `create`, `delete`, `get`, `list`, `patch`, `update` and `watch`.
* `L34Route`: `watch`, `list` and `get`.

### Alternatives for Application Network Configuration Injection

Injecting VIPs and source-based routes into the application is essential to ensure correct network behavior. More details about this process are provided in the dataplane section.

Source-based routes need to be dynamically aligned with the SLLBR instances, ensuring that when SLLBRs are scaled, traffic continues to be routed to the appropriate destinations without disruption. This alignment is crucial to maintain seamless connectivity and load distribution across the infrastructure.

Similarly, VIPs must correspond to those served by the `Gateway`. While not an immediate requirement, VIPs can be added, updated and removed at the same time as the SBRs maintaining consistency with the `Gateway`. Since it is not an immediate requirement, application pods can be recreated when new VIPs are introduced. The responsibilities to add those VIPs can also be delegated on the user side.

There are multiple ways to address this challenge, each offering different levels of automation, flexibility, and complexity. More solutions could also be imagined such as direct server return (DSR) to solve this challenge.

#### Annotation Injection

In this solution, source-based routes and VIPs required by the application are exposed through pod annotations. These annotations represent the desired state and provide visibility into the intended network configuration. This approach allows another component to detect these annotations and inject the necessary configuration directly into the pod ensuring the pod network configuration is adjusted accordingly based on this desired state. 

![Overview](resources/diagrams-Annotation-injection.png)

Different implementation approaches can be considered to achieve annotation-based injection, each with its own trade-offs. Two primary methods are highlighted: one that places the responsibility on the infrastructure, ensuring that network configurations are applied automatically with minimal privileges at the pod level, and another that requires more privileges within the pod itself, offering greater flexibility but increasing security considerations.

##### Dynamic Network Attachment

To inject VIP and SBRs in the pod based on its annotations, this solution leverages Multus-Thick and [Multus-Dynamic-Networks-Controller](https://github.com/k8snetworkplumbingwg/multus-dynamic-networks-controller), which provide the capability to add and reconcile annotations while a pod is running. This allows for dynamic network configuration changes without requiring pod restarts.

![Overview](resources/diagrams-Multus.png)

The approach relies on the [Macvlan](https://www.cni.dev/plugins/current/main/macvlan/) and the [Source-Based Routing (SBR)](https://www.cni.dev/plugins/current/meta/sbr/) CNI plugins. An additional attachment is configured using Macvlan on top of the secondary network interface to handle VIP addresses and source-based routing.

Here is an example below of how the behavior will be when adding a new VIP and when scaling down the SLLBRs.

| Phase 1 | Phase 2 | Phase 3 |
| -------- | ------- | ------- |
| ![Overview](resources/diagrams-MACVLAN-1.png) | ![Overview](resources/diagrams-MACVLAN-2.png) | ![Overview](resources/diagrams-MACVLAN-3.png) |

 In phase 1, the system operates with its current configuration, handling traffic as expected. During phase 2, a new network attachment is introduced to accommodate updated requirements, such as adding a new VIP (`40.0.0.1/32`) and modifying source-based routes to reflect the upcoming removal of SLLR-2. Although both configurations temporarily coexist, they are designed to avoid conflicts. In phase 3, the outdated configuration is removed, achieving the network update without traffic disruption. The only noticeable change is the source MAC address in response packets which will now reflect the newly added macvlan interface.

Here is an example below of a CNI config used to config the VIPs and SBRs on the secondary network interface in an application pod:
```json
{
    "cniVersion": "1.0.0",
    "name": "sbr-vip-pod-a",
    "plugins": [
        {
            "type": "macvlan",
            "master": "net1", // secondary interface in the application pod
            "linkInContainer": true,
            "ipam": {
                "type": "static",
                "addresses": [
                    {
                        "address": "20.0.0.1/32", // VIP
                    },
                ],
                "routes": [
                    { 
                        "dst": "20.0.0.1/32",
                        "table": 5000,
                        "scope": 253, // link
                    },
                    { 
                        "dst": "0.0.0.0/0",
                        "gw": "172.16.0.1", // SLLBR IP
                        "table": 5000,
                    },
                ],
            }
        },
        {
            "type": "tuning",
            "sysctl": {
                "net.ipv4.conf.IFNAME.arp_ignore": "8"
            }
        },
        {
            "type": "sbr",
            "table": 5000
        }
    ]
}
```

In this scenario, managing `NetworkAttachmentDefinition` objects can be complex, requiring careful handling to avoid issues such as stale configurations and resource leakage. Proper coordination and orchestration is essential to prevent conflicts, such as collisions in table IDs, and to ensure seamless updates in response to network changes.

Another potential challenge with this solution is that upon Meridio 2.x uninstallation, network configurations may remain on the application pod. However, they can be easily removed by clearing the corresponding pod annotations.

Adopting this approach also implies that other secondary network providers must support dynamic network attachment, which would require additional development effort for each supported solution.

To perform these operations, a new controller requires access to a set of Kubernetes API resources:
* `Pod`: `patch`, `update` , `watch`, `list` and `get`.

This approach has been demonstrated in the proof-of-concept available here: [LionelJouin/l-3-4-gateway-api-poc](https://github.com/LionelJouin/l-3-4-gateway-api-poc).

##### Network Daemon

In this solution, a DaemonSet is deployed to manage the pod network configuration by reading the pod annotations. The daemon accesses the network namespace of the pod and removes any outdated network configurations and applies new ones based on the annotations of the pod (expected state).

The DaemonSet is deployed alongside Meridio 2.x so it can access the network namespace of the pods across the entire cluster.

Status information about the network configuration is captured and reported within pod annotations by the network daemon. This ensures that the previously applied configurations can be retrieved and referenced if the expected state needs to be updated, re-applied or removed. This approach also provides visibility into the system's configuration over time.

To function properly, the DaemonSet requires access to the Kubernetes API and CRI API to locate and modify the network namespace of the pods. Additionally, specific capabilities (to be clarified) are needed to manipulate the network namespace of the pods.

During uninstallation, residual configurations could be left on the application pod. A job running on each node can be used to clean up any leftover network configurations, ensuring the system is fully cleaned after Meridio 2.x is removed.

This solution is demonstrated in the proof of concept at: [LionelJouin/meridio-2.x](https://github.com/LionelJouin/meridio-2.x).

#### Sidecar

In this solution, a container with specific privileges is running as a sidecar within the application pod to configure source-based routes and VIPs tailored to the services the pod is serving. This sidecar ensures that the networking configuration is dynamically adjusted. 

To perform these operations, the sidecar requires access to a set of Kubernetes API resources (to be defined) and requires the `NET_ADMIN` capability to manage and modify routing tables and IPs.

The sidecar concept was demonstrated in the early phase of Nordix/Meridio: [v0.1.0-alpha](https://github.com/Nordix/Meridio/tree/v0.1.0-alpha)

#### Static Configuration

In this scenario, the user must reserve dedicated internal IPs for allocation to the SLLBRs. Additional configuration is also required on the user side to pre-configure the VIPs and source-based routes towards these dedicated internal IPs. These dedicated internal IPs will be picked by the SLLBRs and configured on their own network interface by themselves.

To ensure proper allocation, the SLLBRs need to sync their states to prevent IP conflicts, ensuring that no IP is assigned to more than one SLLBR. This coordination is vital for maintaining network integrity.

Here is an example below on the behavior when scaling down the SLLBRs:
| Phase 1 | Phase 2 |
| -------- | ------- |
| ![Overview](resources/diagrams-static-1.png) | ![Overview](resources/diagrams-static-2.png) |

In phase 1, SLLBR-1 is assigned the IP `192.2.1.251` and SLLBR-2 is assigned `192.2.1.252`, so the outgoing traffic from the application pod is being routed only between these two pods. In phase 2, when the SLLBRs are scaled down, `192.2.1.251` is released and reassigned to SLLBR-3. 

This has not been demonstrated, there may be issues when scaling SLLBRs in the application pod due to neighbor table cache, potentially causing disruptions in traffic routing.

### Components Nordix/Meridio Differences

Below is a comparison of the communication between components in Nordix/Meridio and Meridio 2.x:
| Nordix/Merdio | Meridio 2.x |
| -------- | ------- |
| ![Overview](resources/diagrams-Communications-v1.png) | ![Overview](resources/diagrams-Communications-v2.png) |

Below is a comparison of the configuration management in Nordix/Meridio and Meridio 2.x:
| Nordix/Merdio | Meridio 2.x |
| -------- | ------- |
| ![Overview](resources/diagrams-Configuration-v1.png) | ![Overview](resources/diagrams-Configuration-v2.png) |

Below is a comparison of the endpoint registration in Nordix/Meridio and Meridio 2.x:
| Nordix/Merdio | Meridio 2.x |
| -------- | ------- |
| ![Overview](resources/diagrams-target-registration-v1.png) | ![Overview](resources/diagrams-target-registration-v2.png) |

## Data Plane

### DC Gateway to Meridio Gateway

The networks connecting the DC Gateway to the Meridio 2.x `Gateway` (SLLBRs) are defined by the user within the `Gateway` specification. The user is then responsible for properly configuring these networks to ensure seamless communication between the DC Gateway and the Meridio 2.x `Gateway`.

### Stateless-Load-Balancer-Router (SLLBR)

The dataplane architecture remains unchanged compared to the previous Nordix/Meridio implementation (refer to the [Nordix/Meridio LB dataplane documentation](https://meridio.nordix.org/docs/dataplane/stateless-lb-frontend)) for further details).

Packets destined for a VIP are redirected to userspace using nftables nfqueues. NFQLB listens on these queues, classifying incoming packets based on `L34Routes` that point to defined `Services`. Each `Service` maintains a list of endpoints, represented within NFQLB by unique identifiers. Once NFQLB processes a packet using the Maglev algorithm to select an endpoint, it assigns the corresponding identifier to the forwarding mark of the packet before returning it to the kernel space.

In the kernel, policy routing leverages the forwarding mark to determine the appropriate routing table and forward the packet to its destination.

For packets originating from a VIP, source-based routing is employed to ensure traffic is directed correctly based on the source IP.

Below is an illustration of this dataplane workflow:

![Overview](resources/diagrams-dataplane-sllbr.png) 

### Meridio Gateway to Endpoint (Application pod)

The responsibility for defining the connection between the Meridio 2.x `Gateway` (SLLBR) and the endpoints lies with the user. This configuration is specified within the application pods of the user and the `Gateway` resource.

The Gateway must be properly connected to the relevant network(s) and reside within the same subnet(s) from which it will select endpoints. It is essential for the user to ensure that the network topology aligns with the intended traffic flow to facilitate proper communication between the Meridio `Gateway` and the endpoints.

### Endpoint (Application pod)

The dataplane architecture remains unchanged compared to the previous Nordix/Meridio implementation (refer to the [Nordix/Meridio target dataplane documentation](https://meridio.nordix.org/docs/dataplane/target)) for further details).

The application pod must handle traffic destined for the VIP as local traffic. To achieve this, the VIP should be assigned as a local IP on an interface within the pod.

If the application needs to accept traffic for an entire CIDR block, the specified CIDR (e.g., `20.0.0.0/24`) must be configured on the loopback (`lo`) interface to ensure proper routing within the pod.

Additionally, all reply traffic and any traffic originating from the application pod with the VIP as the source IP must be routed through the SLLBRs. This ensures that the return path and outbound traffic follow the intended network paths.

### Dataplane Nordix/Meridio Differences

Below is a comparison of the dataplane overview in Nordix/Meridio and Meridio 2.x:
| Nordix/Merdio | Meridio 2.x |
| -------- | ------- |
| ![Overview](resources/diagrams-dataplane-v1.png) | ![Overview](resources/diagrams-dataplane-v2.png) |

## Extra Features

### Port Address Translation (PAT)

TODO

### Gateway Scaling

TODO

## Prerequisites

The solution requires a Kubernetes cluster running at least version 1.28. Alternatively, distributions such as OpenShift are supported, provided they meet the necessary requirements. In some cases, privilege escalation might be required for proper functionality.

Multus is essential to provide secondary network interfaces to pods. Depending on the specific implementation, specific Multus version (Multus-Thick) and additional components such as the Multus-Dynamic-Networks-Controller may also be required.

Furthermore, the Gateway API is a mandatory component to some parts of the API.

## Multi-Tenancy

Meridio 2.x supports multi-tenancy by allowing deployments to be scoped per namespace, with the exception of the Custom Resource Definitions (CRDs), which are shared across the cluster. This approach enables multiple tenants to deploy their own instances of Meridio 2.x, even different versions, within the same Kubernetes cluster without conflicts.

The controllers operate within the boundaries of their respective namespaces, ensuring that resources are reconciled only within the namespace they are deployed in.

It is important to ensure that there are no overlapping network configurations between tenants. However, the responsibility for preventing such conflicts falls on the user, as network isolation and configuration management are outside the scope of this project.

## Upgrade and Migration

### Data Plane

Upgrading Multus does not introduce any traffic disruption, as Multus itself does not carry traffic but rather provides the networking infrastructure to attach multiple network interfaces to pods.

The impact on traffic during an upgrade primarily depends on the Container Network Interface (CNI) plugin in use. Different CNIs may have varying behaviors and potential disruptions during upgrades.

Both the upgrade process for Multus and the CNI plugin fall outside the scope of this project, and users are responsible for planning and executing upgrades in a way that minimizes potential downtime.

### Meridio 2.x

Minimal traffic disturbance, such as packet loss and retransmissions can be expected during the upgrade of the load balancer components.

To reduce the impact on the system, it may be possible to scale the control plane and data plane components, ensuring a smoother transition and minimizing potential disruptions.

All components of Meridio 2.x are stateless, meaning they only reconcile objects and verify that the system state aligns with the specified configuration. If a component fails or is temporarily unavailable, it can be restarted or recreated without loss of state, and it will resume reconciling to restore the desired state.

### From Nordix/Meridio

Meridio 2.x is not backward compatible with Nordix/Meridio, meaning an upgrade without traffic disturbance or configuration changes is not possible. Meridio 2.x does not handle or manage any configuration from Nordix/Meridio, and vice versa.

While both versions can be deployed alongside each other within the same environment, there are no guarantees that source-based routing or any other networking configuration within the application pods will not encounter conflicts.

## Project Structure and Implementation

### Projects

The Meridio 2.x solution can be divided into three distinct projects, in addition to several community-owned projects that it builds upon, such as [k8snetworkplumbingwg/Multus](https://github.com/k8snetworkplumbingwg/multus-cni), [containernetworking/CNI](https://github.com/containernetworking/cni), [kubernetes-sigs/gateway-api](https://github.com/kubernetes-sigs/gateway-api) and the future Kubernetes generic EndpointSlice Controller.

The first project is [Nordix/nfqueue-loadbalancer](https://github.com/Nordix/nfqueue-loadbalancer) (NFQLB), which remains unchanged from its Nordix/Meridio origins. It serves as the core load-balancing component within the SLLB container by leveraging nfqueue.

The second project focuses on the development of a reusable component that integrates with the `GatewayRouter` API and the `Router` container. This component is designed to be adaptable and can be leveraged by similar projects that handle traffic in different ways, such as stateful load-balancers. It remains independent of the load-balancer itself, meaning it could potentially be replaced by an alternative implementation without affecting the overall solution described in this document.

Finally, the Stateless-Load-Balancer project could bring together all these elements to deliver a comprehensive and complete end-to-end solution.

### Framework amd Design Pattern

The [Controller-runtime](https://github.com/kubernetes-sigs/controller-runtime), which has been previously used in Nordix/Meridio, remains a viable option for handling simple use cases. It follows a common Kubernetes pattern that is efficient, easy to develop, and straightforward to test. This makes it an ideal choice for standard reconciliation tasks that do not require deep customization.

For more advanced use cases that demand greater control and performance, a custom informer-based controller can be implemented. This approach (here is an example: [kubernetes/kubernetes/pkg/controller/endpointslice](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/endpointslice/endpointslice_controller.go))) provides fine-grained control over resource handling and offers improved performance at the cost of increased complexity in both implementation and maintenance. It requires a deeper understanding of Kubernetes internals but allows for highly optimized operations.

The upcoming Generic EndpointSlice Controller, as discussed in the [sig-network meeting on 2025-01-16](https://docs.google.com/document/d/1_w77-zG_Xj0zYvEMfQZTQ-wPP4kXkpGD8smVtW_qqWM/edit?tab=t.0), presents an opportunity for managing EndpointSlices efficiently within the system. Leveraging this controller could simplify EndpointSlice management while maintaining optimal performance and scalability.

For API and CRD development, [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) can used to provide built-in schema validation capabilities, which simplifies the process and reduces the need for additional validation mechanisms. By relying on built-in schema validation, the complexity of certificate handling for webhooks can be avoided, leading to a more maintainable and streamlined solution.

## Evolution

### Dynamic Resource Allocation

[kubernetes-sigs/cni-dra-driver](https://github.com/kubernetes-sigs/cni-dra-driver)

Provides Scheduling

newly created KEP allowing the allocation of virtual devices 

[KEP 5075: support dynamic device provisioning](https://github.com/kubernetes/enhancements/issues/5075)

Provides validation [containernetworking/cni#1132](https://github.com/containernetworking/cni/issues/1132)

Provides native integration with the Kubernetes
Status for network interfaces 
merged and part of Kubernetes v1.32
[KEP-4817: Resource Claim Status with possible standardized network interface data](https://github.com/kubernetes/enhancements/issues/4817)

Provides Downward API

### Service Type

Example via a potential Stateful Load-Balancer

This would apply to any other type of implementation providing feature not suiting the stateless-Load-Balancer project.

Accelerate packet processing workload via DPDK

### Service Chaining

supported out-of-the-box

## Alternatives

### LoxiLB

TODO

### OVN-Kubernetes

TODO

### F5

TODO

### Google

TODO

### Cilium

TODO