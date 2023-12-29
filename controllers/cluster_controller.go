/*
Copyright 2023 Dan Molik.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dmolik/argocd-cluster-register/cni/cilium"
	"github.com/dmolik/argocd-cluster-register/conf"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *conf.Config
}

//+kubebuilder:rbac:groups=cluster.argoproj.io,resources=generators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.argoproj.io,resources=generators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cluster.argoproj.io,resources=generators/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status;clusters/finalizers,verbs=get;list;watch
//+kubebuilder:rbac:namespace=argocd,resources=secrets,verbs=create;update;delete;get
//+kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=update;list;watch;get
//+kubebuilder:rbac:resources=secrets,verbs=get,watch,list

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	cluster := capiv1.Cluster{}
	err := r.Get(ctx, req.NamespacedName, &cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.V(0).Info(fmt.Sprintf("found cluster, phase=%s, control_plane_ready=%t", cluster.Status.Phase, cluster.Status.ControlPlaneReady)) // , cluster.Status.Conditions))
	if cluster.Status.Phase == "Deleting" {
		// delete the cluster secret from argocd
		kcfg, err := r.getKubeConfig(ctx, req)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		if _, err = r.deleteSecret(ctx, kcfg); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.removeFromProject(ctx, kcfg); err != nil {
			return ctrl.Result{}, err
		}

	}
	if cluster.Status.Phase != "Deleting" {
		// get the secret and push it into argocd
		kcfg, err := r.getKubeConfig(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if _, err = r.ensureSecret(ctx, kcfg); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.addToProject(ctx, kcfg); err != nil {
			return ctrl.Result{}, err
		}
		if cluster.Status.Phase != "Deleting" {
			if _, err := r.createCNI(ctx, req, cluster); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	oneMinute, err := time.ParseDuration("1m")
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: oneMinute}, nil
}

func (r *ClusterReconciler) createCNI(ctx context.Context, req ctrl.Request, cluster capiv1.Cluster) (ctrl.Result, error) {

	resourceSet := &addonsv1.ClusterResourceSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterResourceSet",
			APIVersion: "addons.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "-cni",
			Namespace: req.Namespace,
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name": req.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: cluster.APIVersion,
					Kind:       cluster.Kind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
		Spec: addonsv1.ClusterResourceSetSpec{
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: cluster.Labels,
			},
			Resources: []addonsv1.ResourceRef{
				{
					Name: req.Name + "-cni",
					Kind: "ConfigMap",
				},
			},
		},
	}
	if err := r.Create(ctx, resourceSet); err != nil {
		return ctrl.Result{}, err
	}

	cni, err := templateClusterCNI(cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "-cni",
			Namespace: req.Namespace,
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name": req.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: cluster.APIVersion,
					Kind:       cluster.Kind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
		Data: map[string]string{
			"cni.yaml": cni,
		},
	}
	if err := r.Create(ctx, cm); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func templateClusterCNI(cluster capiv1.Cluster) (string, error) {
	return cilium.Fetch(cluster.Spec.ControlPlaneEndpoint.Host, cluster.Spec.ControlPlaneEndpoint.Port)
}

func (r *ClusterReconciler) getKubeConfig(ctx context.Context, req ctrl.Request) (*clientcmdapi.Config, error) {
	secret := corev1.Secret{}
	secretReq := req.NamespacedName
	secretReq.Name = secretReq.Name + "-kubeconfig"
	err := r.Get(ctx, secretReq, &secret)
	if err != nil {
		return nil, err
	}
	kubeconfig, err := clientcmd.Load(secret.Data["value"])
	if err != nil {
		return nil, err
	}
	return kubeconfig, nil
}

func (r *ClusterReconciler) deleteSecret(ctx context.Context, kubeconfig *clientcmdapi.Config) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	log.V(0).Info("deleting " + clusterName)
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-cluster-secret",
			Namespace: "argocd",
		},
	}
	err := r.Delete(ctx, &secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) ensureSecret(ctx context.Context, kubeconfig *clientcmdapi.Config) (ctrl.Result, error) {
	clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	authName := kubeconfig.Contexts[kubeconfig.CurrentContext].AuthInfo
	server := kubeconfig.Clusters[clusterName].Server
	config := argoappv1.ClusterConfig{
		TLSClientConfig: argoappv1.TLSClientConfig{
			CAData:   kubeconfig.Clusters[clusterName].CertificateAuthorityData,
			CertData: kubeconfig.AuthInfos[authName].ClientCertificateData,
			KeyData:  kubeconfig.AuthInfos[authName].ClientKeyData,
		},
	}
	if strings.Contains(server, "eks") {
		config.AWSAuthConfig = &argoappv1.AWSAuthConfig{
			ClusterName: clusterName,
			RoleARN:     r.Config.RoleARN,
		}
	}
	configByte, err := json.Marshal(&config)
	if err != nil {
		return ctrl.Result{}, err
	}

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-cluster-secret",
			Namespace: "argocd",
			Labels: map[string]string{
				"app.kubernetes.io/part-of":      "argocd",
				"argocd.argoproj.io/secret-type": "cluster",
				"cluster.x-k8s.io/cluster-name":  clusterName,
			},
		},
		StringData: map[string]string{
			"name":   clusterName,
			"server": server,
			"config": string(configByte),
		},
		Type: "Opaque",
	}
	err = r.Create(ctx, &secret)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			err = r.Update(ctx, &secret)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) removeFromProject(ctx context.Context, kubeconfig *clientcmdapi.Config) error {
	clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	server := kubeconfig.Clusters[clusterName].Server

	for _, proj := range r.Config.Projects {
		project := argoappv1.AppProject{}
		projectReq := types.NamespacedName{
			Name:      proj,
			Namespace: "argocd",
		}
		err := r.Get(ctx, projectReq, &project)
		if err != nil {
			return err
		}
		for idx, dest := range project.Spec.Destinations {
			if dest.Name == clusterName {
				project.Spec.Destinations = append(project.Spec.Destinations[:idx], project.Spec.Destinations[idx+1:]...)
				break
			}
			if dest.Server == server {
				project.Spec.Destinations = append(project.Spec.Destinations[:idx], project.Spec.Destinations[idx+1:]...)
				break
			}
		}
		err = r.Update(ctx, &project)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ClusterReconciler) addToProject(ctx context.Context, kubeconfig *clientcmdapi.Config) error {
	clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	server := kubeconfig.Clusters[clusterName].Server
	for _, proj := range r.Config.Projects {
		project := argoappv1.AppProject{}
		projectReq := types.NamespacedName{
			Name:      proj,
			Namespace: "argocd",
		}
		err := r.Get(ctx, projectReq, &project)
		if err != nil {
			return err
		}
		for _, dest := range project.Spec.Destinations {
			if dest.Name == clusterName {
				return nil
			}
			if dest.Server == server {
				return nil
			}
		}
		project.Spec.Destinations = append(project.Spec.Destinations, argoappv1.ApplicationDestination{
			Name:   clusterName,
			Server: server,
		})
		err = r.Update(ctx, &project)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capiv1.Cluster{}).
		Complete(r)
}
