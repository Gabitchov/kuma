# Developer documentation

## Pre-requirements

- `curl`
- `git`
- `go`
- `make`

For a quick start, use the official `golang` Docker image (which has all these tools pre-installed), e.g.

```bash
docker run --rm -ti \
  --user 65534:65534 \
  --volume `pwd`:/go/src/github.com/Kong/konvoy/components/konvoy-control-plane \
  --workdir /go/src/github.com/Kong/konvoy/components/konvoy-control-plane \
  --env HOME=/tmp/home \
  --env GO111MODULE=on \
  golang:1.12.5 bash
export PATH=$HOME/bin:$PATH
```

## Helper commands

```bash
make help
```

## Installing dev tools

Run:

```bash
make dev/tools
```

which will install the following tools at `$HOME/bin`:

1. [Ginkgo](https://github.com/onsi/ginkgo#set-me-up) (BDD testing framework)
2. [Kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) (Kubernetes API extension framework, comes with `etcd` and `kube-apiserver`)
3. [kustomize](https://book.kubebuilder.io/quick-start.html#installation) (Customization of kubernetes YAML configurations)
4. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-binary-with-curl-on-linux) (Kubernetes API client)
5. [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes IN Docker)
6. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/#linux) (Kubernetes in VM)

ATTENTION: By default, development tools will be installed at `$HOME/bin`. Remember to include this directory into your `PATH`, 
e.g. by adding `export PATH=$HOME/bin:$PATH` line to the `$HOME/.bashrc` file.

## Building

Run:

```bash
make build
```

## Integration tests

 Integration tests will run all dependencies (ex. Postgres). Run:

 ```bash
make integration
```

## Running Control Plane on local machine

### Universal without any external dependency

1. Run `Control Plane` on local machine:

```bash
make run/universal/memory
```

### Standalone with Postgres as a storage

1. Run Postgres with initial schema using docker-compose.
It will run on port 15432 with username: `konvoy`, password: `konvoy` and db name: `konvoy`.

```bash
make start/postgres
```

2. Run `Control Plane` on local machine.

```bash
make run/universal/postgres
```

This will also start

### Kubernetes

1. Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/k8s

# set KUBECONFIG for use by `konvoyctl` and `kubectl`
export KUBECONFIG="$(kind get kubeconfig-path --name=konvoy)"
```

2. Run `Control Plane` on local machine:

```bash
make run/k8s
```

### Check the setup

Note: for the moment. Only K8S setup passes checks.

1. Make a test `Discovery` request to `LDS`:

```bash
make curl/listeners
```

2. Make a test `Discovery` request to `CDS`:

```bash
make curl/clusters
```

## Pointing Envoy at Control Plane

1. Start the Control Plane in your preferable choice described above 

2. Build `konvoy-dataplane`

```bash
make build/konvoy-dp

export PATH=`pwd`/build/artifacts/konvoy-dataplane:$PATH
```

3. Start `Envoy` on local machine (requires `envoy` binary to be on your `PATH`):

```bash
make run/example/envoy
```

4. Dump effective `Envoy` config:

```bash
make config_dump/example/envoy
```

## Running Control Plane on Kubernetes

1. Run [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/k8s

# set KUBECONFIG for use by `konvoyctl` and `kubectl`
export KUBECONFIG="$(kind get kubeconfig-path --name=konvoy)"
```

2. Deploy `Control Plane` to [KIND](https://kind.sigs.k8s.io/docs/user/quick-start) (Kubernetes IN Docker):

```bash
make start/control-plane/k8s
```

3. Redeploy demo app (to get Konvoy sidecar injected)

```bash
kubectl delete -n konvoy-demo pod -l app=demo-app
```

4. Build `konvoyctl`

```bash
make build/konvoyctl

export PATH=`pwd`/build/artifacts/konvoyctl:$PATH
```

4. Add `Control Plane` to your `konvoyctl` config:

```bash
konvoyctl config control-planes add k8s --name demo
```

5. Verify that `Control Plane` has been added:

```bash
konvoyctl config control-planes list

NAME                      ENVIRONMENT
kubernetes-admin@konvoy   k8s
```

6. List `Dataplanes` connected to the `Control Plane`:

```bash
konvoyctl get dataplanes

MESH      NAMESPACE   NAME                        SUBSCRIPTIONS   LAST CONNECTED AGO   TOTAL UPDATES   TOTAL ERRORS
default               demo-app-685444477b-dnx9t   1               21m9s                2               0
```
