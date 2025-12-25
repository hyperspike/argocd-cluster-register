# ArgoCD-Cluster-Register

## Project Overview

**ArgoCD-Cluster-Register** is a Kubernetes operator (controller) built with Go and Kubebuilder. Its primary purpose is to automate the registration of [Cluster-API (CAPI)](https://cluster-api.sigs.k8s.io/) clusters with [ArgoCD](https://argo-cd.readthedocs.io/).

Instead of ArgoCD or CAPI communicating directly, this controller acts as a bridge:
1.  **Watches** for CAPI `Cluster` resources.
2.  **Registers** them with ArgoCD by creating the necessary Secrets in the `argocd` namespace.
3.  **Updates** specified ArgoCD `AppProject` resources to allow deployments to the new cluster.
4.  **Installs** CNI (Cilium) on the new clusters using `ClusterResourceSet` when the control plane is ready.

### Architecture

*   **Language:** Go (v1.25+)
*   **Framework:** [Kubebuilder](https://book.kubebuilder.io/) / [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
*   **Key APIs Watched:**
    *   `cluster.x-k8s.io/v1beta1` (Cluster API)
*   **Key Resources Managed:**
    *   `v1/Secret` (in `argocd` namespace)
    *   `argoproj.io/v1alpha1/AppProject` (ArgoCD)
    *   `addons.cluster.x-k8s.io/v1beta1/ClusterResourceSet` (for CNI installation)
    *   `v1/ConfigMap` (containing CNI manifests)

## Building and Running

The project uses a `Makefile` to manage build and development tasks.

### Key Commands

*   **Build Binary:**
    ```bash
    make build
    ```
    Creates the `manager` binary in `bin/`.

*   **Run Locally:**
    ```bash
    make run
    ```
    Runs the controller locally, connecting to the cluster defined in `~/.kube/config`.

*   **Run Tests:**
    ```bash
    make test
    ```
    Runs unit tests using `ginkgo` / `gomega`. Requires `envtest` (setup automatically via make).

*   **Build Docker Image:**
    ```bash
    make docker-build
    ```

*   **Push Docker Image:**
    ```bash
    make docker-push
    ```

*   **Generate Manifests:**
    ```bash
    make manifests
    ```
    Regenerates CRD manifests and RBAC YAMLs using `controller-gen`.

*   **Generate Code:**
    ```bash
    make generate
    ```
    Regenerates boilerplate code (DeepCopy methods, etc.).

## Development Conventions

*   **Controller Logic:** The core logic resides in `controllers/cluster_controller.go`.
*   **Configuration:**
    *   Configuration is handled via environment variables (processed in `conf/conf.go`).
    *   `ROLE_ARN`: AWS Role ARN (for EKS clusters).
    *   `PROJECT`: Comma-separated list of ArgoCD project names to add the cluster to.
*   **CNI Management:** CNI (Cilium) templates are located in `cni/cilium/`.
*   **Logging:** Uses `go.uber.org/zap` for logging.
*   **Linting:** `make lint` runs `golangci-lint`.

## Directory Structure

*   `bin/`: Compiled binaries and build tools (e.g., `controller-gen`, `kustomize`).
*   `cni/`: Contains logic for generating CNI manifests (Cilium).
*   `conf/`: Configuration parsing logic.
*   `config/`: Kustomize configuration for launching the controller on a cluster.
*   `controllers/`: The controller implementations (Reconcile loops).
*   `hack/`: Scripts and boilerplate headers.
*   `main/`: The entry point (`main.go`).
