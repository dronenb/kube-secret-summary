/*
Copyright 2025.

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

package controller

import (
	"context"
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	secretsv1alpha1 "github.com/dronenb/kube-secret-summary/api/v1alpha1"
)

// SecretSummaryReconciler reconciles a SecretSummary object
type SecretSummaryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=secrets.k8s.bendronen.com,resources=secretsummaries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secrets.k8s.bendronen.com,resources=secretsummaries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=secrets.k8s.bendronen.com,resources=secretsummaries/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSummary object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *SecretSummaryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the Secret
	var secret v1.Secret
	if err := r.Get(ctx, req.NamespacedName, &secret); err != nil {
		if errors.IsNotFound(err) {
			// If Secret is deleted, delete the corresponding SecretSummary
			log.Info("Secret not found, deleting SecretSummary")

			secretSummary := &secretsv1alpha1.SecretSummary{}
			if err := r.Get(ctx, req.NamespacedName, secretSummary); err == nil {
				if delErr := r.Delete(ctx, secretSummary); delErr != nil {
					log.Error(delErr, "Failed to delete SecretSummary")
					return ctrl.Result{}, delErr
				}
				log.Info("Deleted SecretSummary", "SecretSummary", req.NamespacedName)
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Secret")
		return ctrl.Result{}, err
	}

	// Extract keys from secret data
	keys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		keys = append(keys, key)
	}

	// Create or update the SecretSummary
	secretSummary := &secretsv1alpha1.SecretSummary{}
	err := r.Get(ctx, req.NamespacedName, secretSummary)
	if err != nil && errors.IsNotFound(err) {
		// Create new SecretSummary
		secretSummary = &secretsv1alpha1.SecretSummary{
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.Name,
				Namespace:   secret.Namespace,
				Labels:      secret.Labels,
				Annotations: secret.Annotations,
			},
		}
		if err := r.Create(ctx, secretSummary); err != nil {
			log.Error(err, "Failed to create SecretSummary")
			return ctrl.Result{}, err
		}
		log.Info("Created SecretSummary", "SecretSummary", secretSummary.Name)
	} else if err != nil {
		log.Error(err, "Failed to get SecretSummary")
		return ctrl.Result{}, err
	}

	// Update labels and annotations if they have changed
	needsUpdate := false
	if !reflect.DeepEqual(secretSummary.Labels, secret.Labels) {
		secretSummary.Labels = secret.Labels
		needsUpdate = true
	}
	if !reflect.DeepEqual(secretSummary.Annotations, secret.Annotations) {
		secretSummary.Annotations = secret.Annotations
		needsUpdate = true
	}

	if needsUpdate {
		if err := r.Update(ctx, secretSummary); err != nil {
			log.Error(err, "Failed to update SecretSummary metadata")
			return ctrl.Result{}, err
		}
		log.Info("Updated SecretSummary metadata", "SecretSummary", secretSummary.Name)
	}

	secretSummary.Status.Keys = keys
	secretSummary.Status.KeyCount = len(keys)
	secretSummary.Status.LastUpdated = metav1.Now()
	secretSummary.Status.Type = string(secret.Type)

	if err := r.Status().Update(ctx, secretSummary); err != nil {
		log.Error(err, "Failed to update SecretSummary status")
		return ctrl.Result{}, err
	}

	log.Info("Updated SecretSummary status", "SecretSummary", secretSummary.Name)

	// Requeue in case of updates
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretSummaryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		Complete(r)
}
