# ArgoCD-Cluster-Register

[![scan workflow](https://github.com/dmolik/argocd-cluster-register/actions/workflows/scan.yaml/badge.svg)](https://github.com/dmolik/argocd-cluster-register/actions/workflows/scan.yaml)
[![license](https://badgen.net/github/license/dmolik/argocd-cluster-register/)](https://github.com/dmolik/argocd-cluster-register/blob/main/LICENSE)
[![release](https://badgen.net/github/release/dmolik/argocd-cluster-register/stable)](https://github.com/argocd-cluster-register/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmolik/argocd-cluster-register)](https://goreportcard.com/report/github.com/dmolik/argocd-cluster-register)

An ArgoCD controller to listen for Cluster-API clusters and register them with an ArgoCD project

## How it works

the Cluster-Register controller listens to Kubernetes-API for the [Cluster-API]() Resource [Cluster]() and if a cluster is in a non-deleting state it will search for Cluster Connecting resources and add a deterministic Cluster secret into the ArgoCD namespace. Furthermore it will then add the Cluster to the appropriate ArgoCD [Projects]().

Thus the Cluster-Register controller never contacts CAPI or ArgoCD directly. Providing two benefits, re-use of [Kubernetes RBAC]() and ease of programming as there is only the controller-runtime/kubebuilder to interacte with.

## Getting Started

Please note; ArgoCD-Cluster-Register is still work in progress, and the deployment config is undergoing some updates.

    kubectl apply -k github.com/dmolik/argocd-cluster-register/config/default

### Notes

ArgoCD-Cluster-Register doesn't provide label based filtering at this time, but this feature is planed.

Testing includes [Kubeadm]() and [EKS]() clusters, other declarative auth mechanisms have not been implemented.
