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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	//"k8s.io/client/kubernetes/config/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	// registryv1alpha1 "github.com/dmolik/argocd-cluster-register/api/v1alpha1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=registry.argoproj.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=registry.argoproj.io,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=registry.argoproj.io,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	cluster := v1beta1.Cluster{}

	err := r.Get(ctx, req.NamespacedName, &cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	//log.V(0).Info(fmt.Sprintf("%s: %+v\n", cluster.ObjectMeta.Name, cluster.Status))
	log.V(0).Info(fmt.Sprintf("found cluster, phase=%s, control_plane_ready=%t", cluster.Status.Phase, cluster.Status.ControlPlaneReady)) // , cluster.Status.Conditions))
	if cluster.Status.Phase == "Deleting" {
		return r.getSecret(ctx, req)
		// delete the cluster secret from argocd
		//return ctrl.Result{}, nil
	}
	if cluster.Status.Phase != "Deleting" {
		// get the secret and push it into argocd
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) getSecret(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	secret := corev1.Secret{}
	secretReq := req.NamespacedName
	secretReq.Name = secretReq.Name + "-kubeconfig"
	err := r.Get(ctx, secretReq, &secret)
	if err != nil {
		return ctrl.Result{}, err
	}
	kubeconfig, err := clientcmd.Load(secret.Data["value"])
	if err != nil {
		return ctrl.Result{}, err
	}
	log.V(0).Info(fmt.Sprintf("got secret %+v", kubeconfig))
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) ensureSecret(ctx context.Context, kubeconfig clientcmdapi.Config) (ctrl.Result, error) {

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "argocd",
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		StringData: map[string]string{},
		Type:       "Opaque",
	}
	_ = r.Create(ctx, &secret)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Cluster{}).
		Complete(r)
}
