/*
Copyright 2022 Dan Molik.

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

	clusterv1alpha1 "github.com/dmolik/argocd-cluster-register/api/v1alpha1"

	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// GeneratorReconciler reconciles a Generator object
type GeneratorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cluster.argoproj.io,resources=generators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.argoproj.io,resources=generators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cluster.argoproj.io,resources=generators/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status;clusters/finalizers,verbs=get;list;watch
//+kubebuilder:rbac:namespace=argocd,resources=secrets,verbs=create;update;delete;get
//+kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=update;list;watch;get

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *GeneratorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	gen := clusterv1alpha1.Generator{}
	err := r.Get(ctx, req.NamespacedName, &gen)
	if err != nil {
		return ctrl.Result{}, err
	}

	clusterList := &capiv1beta1.ClusterList{}
	err = r.List(ctx, clusterList, client.MatchingLabels{})
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, cluster := range clusterList.Items {
		log.V(0).Info(fmt.Sprintf("found cluster, phase=%s, control_plane_ready=%t, revision=%s, name=%s", cluster.Status.Phase, cluster.Status.ControlPlaneReady, cluster.ResourceVersion, cluster.ObjectMeta.Name)) // , cluster.Status.Conditions))
		if cluster.Status.Phase == "Deleting" {
			// delete the cluster secret from argocd
			kcfg, err := r.getKubeConfig(ctx, &cluster)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return ctrl.Result{}, err
			}
			if _, err = r.deleteSecret(ctx, kcfg); err != nil {
				return ctrl.Result{}, err
			}
		}
		if cluster.Status.Phase != "Deleting" {
			// get the secret and push it into argocd
			kcfg, err := r.getKubeConfig(ctx, &cluster)
			if err != nil {
				return ctrl.Result{}, err
			}
			if _, err = r.ensureSecret(ctx, kcfg, &cluster); err != nil {
				return ctrl.Result{}, err
			}
			if err = r.addToProject(ctx, kcfg, &gen); err != nil {
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

func (r *GeneratorReconciler) getKubeConfig(ctx context.Context, cluster *capiv1beta1.Cluster) (*clientcmdapi.Config, error) {
	secret := corev1.Secret{}
	secretReq := types.NamespacedName{}
	secretReq.Name = cluster.ObjectMeta.Name + "-kubeconfig"
	secretReq.Namespace = cluster.ObjectMeta.Namespace
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

func (r *GeneratorReconciler) deleteSecret(ctx context.Context, kubeconfig *clientcmdapi.Config) (ctrl.Result, error) {
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

func (r *GeneratorReconciler) ensureSecret(ctx context.Context, kubeconfig *clientcmdapi.Config, cluster *capiv1beta1.Cluster) (ctrl.Result, error) {
	clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	authName := kubeconfig.Contexts[kubeconfig.CurrentContext].AuthInfo
	config := argoappv1.ClusterConfig{
		TLSClientConfig: argoappv1.TLSClientConfig{
			CAData:   kubeconfig.Clusters[clusterName].CertificateAuthorityData,
			CertData: kubeconfig.AuthInfos[authName].ClientCertificateData,
			KeyData:  kubeconfig.AuthInfos[authName].ClientKeyData,
		},
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
			Annotations: map[string]string{
				"cluster.x-k8s.io/revision": cluster.ResourceVersion,
			},
		},
		StringData: map[string]string{
			"name":   clusterName,
			"server": kubeconfig.Clusters[clusterName].Server,
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
func (r *GeneratorReconciler) addToProject(ctx context.Context, kubeconfig *clientcmdapi.Config, gen *clusterv1alpha1.Generator) error {
	clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
	if gen.Spec.AppProjectName == "" {
		return nil
	}
	project := argoappv1.AppProject{}
	projectReq := types.NamespacedName{
		Name:      gen.ObjectMeta.Name,
		Namespace: gen.ObjectMeta.Namespace,
	}
	err := r.Get(ctx, projectReq, &project)
	if err != nil {
		return err
	}
	project.Spec.Destinations = append(project.Spec.Destinations, argoappv1.ApplicationDestination{
		Name: clusterName,
	})
	return r.Update(ctx, &project)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GeneratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha1.Generator{}).
		Complete(r)
}
