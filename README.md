# Automent

[![scan workflow](https://github.com/dmolik/automent/actions/workflows/scan.yaml/badge.svg)](https://github.com/dmolik/automent/actions/workflows/scan.yaml)
[![license](https://badgen.net/github/license/dmolik/automent/)](https://github.com/dmolik/automent/blob/main/LICENSE)
[![release](https://badgen.net/github/release/dmolik/automent/stable)](https://github.com/dmolik/automent/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmolik/automent)](https://goreportcard.com/report/github.com/dmolik/automent)

_Automent_ (*Automatic-Cement*) An ArgoCD controller to listen for Cluster-API clusters and register them with an ArgoCD project

## How it works

the Cluster-Register controller listens to Kubernetes-API for the [Cluster-API](https://cluster-api.sigs.k8s.io/) Resource [Cluster](https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api/cluster.x-k8s.io/Cluster/v1beta1@v1.2.0) and if a cluster is in a non-deleting state it will search for Cluster Connecting resources and add a deterministic Cluster secret into the ArgoCD namespace. Furthermore it will then add the Cluster to the appropriate ArgoCD [Projects](https://doc.crds.dev/github.com/argoproj/argo-cd/argoproj.io/AppProject/v1alpha1@v2.4.4).

Thus the Cluster-Register controller never contacts CAPI or ArgoCD directly. Providing two benefits, re-use of [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) and ease of programming as there is only the controller-runtime/kubebuilder to interacte with.

## Getting Started

Please note; ArgoCD-Cluster-Register is still work in progress, and the deployment config is undergoing some updates.

    kubectl apply -k github.com/dmolik/automent/config/default

### Notes

ArgoCD-Cluster-Register doesn't provide label based filtering at this time, but this feature is planed.

Testing includes [Kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/)/[CAPI-Docs](https://cluster-api.sigs.k8s.io/tasks/kubeadm-bootstrap.html) and [EKS](https://www.amazonaws.cn/en/eks/)/[CAPI-Docs](https://cluster-api-aws.sigs.k8s.io/topics/eks/enabling.html) Clusters, other [declarative](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#clusters) auth mechanisms have not been implemented.
