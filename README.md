# field-exporter
[![dh](./img/dh-logo.png)](#)

## Description
This controller is used to fill the gap
in [k8s-config-connector](https://github.com/GoogleCloudPlatform/k8s-config-connector) for exporting value from Config
Connector managed resources into Secrets and ConfigMaps.

Currently supported Config Connector resources:

- [RedisInstance](https://cloud.google.com/config-connector/docs/reference/resource-docs/redis/redisinstance)

Here is an example resource for the controller:

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

Which will create a `ConfigMap` that can be used to [add environment variables](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables) to your Kubernetes pod:

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
