# field-exporter
[![dh](./img/dh-logo.png)](#)

## Description
This controller is used to fill the gap
in [k8s-config-connector](https://github.com/GoogleCloudPlatform/k8s-config-connector) (KCC) and [AWS Controllers for Kubernetes](https://github.com/aws-controllers-k8s/community) (ACK) for exporting values from managed resources into Secrets and ConfigMaps.

## Supported Resources and Examples

The controller discovers available API resources within the following supported API groups.

### Google Cloud Config Connector (KCC)

- `alloydb.cnrm.cloud.google.com`
- `iam.cnrm.cloud.google.com`
- `redis.cnrm.cloud.google.com`
- `sql.cnrm.cloud.google.com`
- `storage.cnrm.cloud.google.com`

#### KCC Example

Here is an example of exporting fields from a KCC `RedisInstance` to a `ConfigMap`.

```yaml
apiVersion: gdp.deliveryhero.io/v1alpha1
kind: ResourceFieldExport
metadata:
  name: myapp-redis
spec:
  from:
    apiVersion: redis.cnrm.cloud.google.com/v1beta1
    kind: RedisInstance
    name: myapp-redis
  outputs:
    - key: endpoint
      path: .status.host
    - key: port
      path: .status.port
  requiredFields:
    statusConditions:
      - status: "True"
        type: Ready
  to:
    name: myapp-redis-config
    type: ConfigMap
```

Which will update a `ConfigMap` with data that can be used to [add environment variables](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables) to your Kubernetes pod:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: special-config
  namespace: default
data:
  endpoint: 10.111.1.3
  port: 6379
```

The controller can also update a `Secret`.

### AWS Controllers for Kubernetes (ACK)

- `rds.services.k8s.aws`
- `dynamodb.services.k8s.aws`
- `elasticache.services.k8s.aws (for Redis)`

#### ACK Example

For ACK resources, it is recommended to wait for the resource to be synced before exporting fields. This can be achieved by requiring the `ACK.ResourceSynced` condition to be `True`.

Here is an example of exporting the endpoint from an rds.services.k8s.aws DBCluster into a Secret:

```yaml
apiVersion: gdp.deliveryhero.io/v1alpha1
kind: ResourceFieldExport
metadata:
  name: myapp-db-aws
spec:
  from:
    apiVersion: rds.services.k8s.aws/v1alpha1
    kind: DBCluster
    name: myapp-db-aws
  outputs:
  - key: endpoint
    path: .status.endpoint
  - key: reader-endpoint
    path: .status.readerEndpoint
  requiredFields:
    statusConditions:
    - status: "True"
      type: ACK.ResourceSynced
  to:
    name: myapp-db-aws-output
    type: Secret
```

The controller will create or update a Secret with the exported data. The values will be base64 encoded as is standard for Secrets. This can then be consumed by your pods. As shown in the KCC example, the controller can also write to a ConfigMap.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Installation via Helm Chart

Follow [chart doc](https://github.com/deliveryhero/helm-charts/tree/master/stable/field-exporter) to install Field Exporter with CRDs

### Running on the cluster manually

1. Install Instances of Custom Resources:

```sh
kubectl apply -k config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/field-exporter:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/field-exporter:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Contributing

To contribute, please read our [contributing docs](CONTRIBUTING.md).

## License

Copyright © 2023 Delivery Hero SE

Contents of this repository is licensed under the Apache-2.0 [License](LICENSE).
