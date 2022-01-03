/*
Copyright 2021.

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

package commerce

import (
	"context"
	"fmt"
	"regexp"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
	"github.com/k8scommerce/cluster-operator/controllers/constant"
)

// CommerceReconciler reconciles a Commerce object
type CommerceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

var commerceFinalizer = "cache.commerce.k8scommerce.com/finalizer"

//+kubebuilder:rbac:groups=cache.commerce.k8scommerce.com,resources=commerces,verbs=*
//+kubebuilder:rbac:groups=cache.commerce.k8scommerce.com,resources=commerces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.commerce.k8scommerce.com,resources=commerces/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
//+kubebuilder:rbac:groups=core,resources=services,verbs=*
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=*
//+kubebuilder:rbac:groups=*,resources=namespaces,verbs=*
//+kubebuilder:rbac:groups=*,resources=pods,verbs=*
//+kubebuilder:rbac:groups=*,resources=deployments,verbs=*
//+kubebuilder:rbac:groups=*,resources=configmaps;secrets,verbs=*
//+kubebuilder:rbac:groups=*,resources=events,verbs=*
//+kubebuilder:rbac:groups=*,resources=endpoints,verbs=get;watch;list

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CommerceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("commerce", req.NamespacedName)

	var result reconcile.Result

	// check the namespace exists first
	commerce := &cachev1alpha1.Commerce{}
	err := r.Get(ctx, req.NamespacedName, commerce)
	if err != nil {

		// log.Info("ERROR", err.Error(), req.NamespacedName)

		if errors.IsNotFound(err) {
			log.Info("Commerce resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}

	// if we're not being delete add the finalizer
	if !commerce.IsBeingDeleted() {
		// only add the finalizer if it doesn't already exist
		if !commerce.HasFinalizer(commerceFinalizer) {
			r.Log.Info(fmt.Sprintf("AddFinalizer for %v", req.NamespacedName))
			if err := r.addFinalizer(ctx, commerce, log); err != nil {
				return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %w", err)
			}
		}

		// Is this a new generation?
		// isNewGeneration := false
		if commerce.Generation != commerce.Status.ObservedGeneration {
			// isNewGeneration = true
			patch := client.MergeFrom(commerce.DeepCopy())
			commerce.Status.ObservedGeneration = commerce.Generation
			if err = r.Status().Patch(ctx, commerce, patch); err != nil {
				r.Log.Error(err, "Failed to update Deployment status during Reconcile")
				return ctrl.Result{}, err
			}
		}

		// // Create and/or sync the Secrets from yaml to a Secret
		// if requeue, err := r.secretsSync(ctx, deploymentFound, isNewGeneration); err != nil || requeue {
		// 	return ctrl.Result{Requeue: requeue}, err
		// }

		// // Reconcile environment variables.
		// if requeue, err := r.environmentVariablesSync(ctx, deploymentFound); err != nil || requeue {
		// 	return ctrl.Result{Requeue: requeue}, err
		// }

		// add the services, deployments, pods, etc...
		if constant.TargetNamespace != "" {
			// Reconcile Target Namespace
			result, err = r.reconcileNamespace(ctx, commerce, log)
			if err != nil {
				return result, err
			}

			// Reconcile Etc
			// result, err = r.reconcileEtcd(ctx, commerce, log)
			// if err != nil {
			// 	return result, err
			// }
		}

		// GatewayClient
		if commerce.Spec.CoreMicroServices.GatewayClient != nil {
			// Reconcile GatwayClient object
			result, err = r.reconcileDeployment(ctx, commerce, NewGatewayClient(), log)
			if err != nil {
				return result, err
			}

			// Reconcile GatwayService object
			result, err = r.reconcileGatewayClientService(ctx, commerce, log)
			if err != nil {
				return result, err
			}

			// Reconcile GatwayIngress object
			result, err = r.reconcileGatewayClientIngress(ctx, commerce, log)
			if err != nil {
				return result, err
			}
		}

		// GatewayAdmin
		if commerce.Spec.CoreMicroServices.GatewayAdmin != nil {
			// Reconcile GatwayAdmin object
			result, err = r.reconcileDeployment(ctx, commerce, NewGatewayAdmin(), log)
			if err != nil {
				return result, err
			}

			// Reconcile GatwayService object
			result, err = r.reconcileGatewayAdminService(ctx, commerce, log)
			if err != nil {
				return result, err
			}

			// Reconcile GatwayIngress object
			result, err = r.reconcileGatewayAdminIngress(ctx, commerce, log)
			if err != nil {
				return result, err
			}
		}

		// Cart
		if commerce.Spec.CoreMicroServices.Cart != nil {
			// Reconcile Cart object
			result, err = r.reconcileDeployment(ctx, commerce, NewCart(), log)
			if err != nil {
				return result, err
			}
		}

		// Customer
		if commerce.Spec.CoreMicroServices.Customer != nil {
			// Reconcile Customer object
			result, err = r.reconcileDeployment(ctx, commerce, NewCustomer(), log)
			if err != nil {
				return result, err
			}
		}

		// Email
		if commerce.Spec.CoreMicroServices.Email != nil {
			// Reconcile Email object
			result, err = r.reconcileDeployment(ctx, commerce, NewEmail(), log)
			if err != nil {
				return result, err
			}
		}

		// Inventory
		if commerce.Spec.CoreMicroServices.Inventory != nil {
			// Reconcile Inventory object
			result, err = r.reconcileDeployment(ctx, commerce, NewInventory(), log)
			if err != nil {
				return result, err
			}
		}

		// OthersBought
		if commerce.Spec.CoreMicroServices.OthersBought != nil {
			// Reconcile OthersBought object
			result, err = r.reconcileDeployment(ctx, commerce, NewOthersBought(), log)
			if err != nil {
				return result, err
			}
		}

		// Payment
		if commerce.Spec.CoreMicroServices.Payment != nil {
			// Reconcile Payment object
			result, err = r.reconcileDeployment(ctx, commerce, NewPayment(), log)
			if err != nil {
				return result, err
			}
		}

		// Product
		if commerce.Spec.CoreMicroServices.Product != nil {
			// Reconcile Product object
			result, err = r.reconcileDeployment(ctx, commerce, NewProduct(), log)
			if err != nil {
				return result, err
			}
		}

		// Shipping
		if commerce.Spec.CoreMicroServices.Shipping != nil {
			// Reconcile Shipping object
			result, err = r.reconcileDeployment(ctx, commerce, NewShipping(), log)
			if err != nil {
				return result, err
			}
		}

		// SimilarProducts
		if commerce.Spec.CoreMicroServices.SimilarProducts != nil {
			// Reconcile SimilarProducts object
			result, err = r.reconcileDeployment(ctx, commerce, NewSimilarProducts(), log)
			if err != nil {
				return result, err
			}
		}

		// Store
		if commerce.Spec.CoreMicroServices.Store != nil {
			// Reconcile Store object
			result, err = r.reconcileDeployment(ctx, commerce, NewStore(), log)
			if err != nil {
				return result, err
			}
		}

		// User
		if commerce.Spec.CoreMicroServices.User != nil {
			// Reconcile User object
			result, err = r.reconcileDeployment(ctx, commerce, NewUser(), log)
			if err != nil {
				return result, err
			}
		}

		// Warehouse
		if commerce.Spec.CoreMicroServices.Warehouse != nil {
			// Reconcile Warehouse object
			result, err = r.reconcileDeployment(ctx, commerce, NewWarehouse(), log)
			if err != nil {
				return result, err
			}
		}

	} else {
		if err := r.handleFinalizer(ctx, commerce, log); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %w", err)
		}
	}

	// return ctrl.Result{}, nil
	return ctrl.Result{RequeueAfter: constant.ReconcileRequeueAfter}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CommerceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Commerce{}).
		Complete(r)
}

// NewInt32 converts int32 to a pointer.
func NewInt32(val int32) *int32 {
	p := new(int32)
	*p = val
	return p
}

// NewInt64 converts int64 to a pointer.
func NewInt64(val int64) *int64 {
	p := new(int64)
	*p = val
	return p
}

func statusMessageFromReplicas(len int, size int32) string {
	if int32(len) != size {
		return fmt.Sprintf("Size Issue: (%d,%d)  %s", size, len, time.Now())
	}
	return fmt.Sprintf("Size Match: (%d,%d)  %s", size, len, time.Now())
}

// getRunningPodNames returns the pod names for the pods running in the array of pods passed in.
func getRunningPodNames(pods []corev1.Pod) []string {
	// Create an empty []string, so if no podNames are returned, instead of nil we get an empty slice
	var podNames []string = make([]string, 0)
	for _, pod := range pods {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			podNames = append(podNames, pod.Name)
		}
	}
	return podNames
}

// CleanContainerImage removes beginning http:// or https://
func CleanContainerImage(image string) string {
	re := regexp.MustCompile(`^https?://`)
	cleanedImage := re.ReplaceAllString(image, "")
	return cleanedImage
}

func (r *CommerceReconciler) findComponentsByLabel(namespace string, matchingLabels client.MatchingLabels) (*v1.PodList, error) {
	podList := &v1.PodList{}

	opts := []client.ListOption{
		client.InNamespace(namespace),
		matchingLabels,
	}

	ctx := context.Background()
	err := r.List(ctx, podList, opts...)
	return podList, err
}

func (r *CommerceReconciler) getRunningEtcdPods(cr *cachev1alpha1.Commerce) int32 {
	return 3
	// var running int32 = 0

	// podList, err := r.findComponentsByLabel(constant.TargetNamespace, client.MatchingLabels{
	// 	"app": "etcd",
	// })

	// if err != nil {
	// 	r.Log.Error(err, "podlist")
	// }

	// for _, pod := range podList.Items {
	// 	if pod.Status.Phase == v1.PodRunning {
	// 		running++
	// 		// r.Log.Info("EtcdPod", "Pod is running", running, "pod.Status", pod.Status)
	// 	}
	// }

	// return running
}

func ensureResourceDefaults(cr *cachev1alpha1.Commerce, microService *cachev1alpha1.MicroService) {
	if cr.Spec.CoreMicroServices.GatewayClient != nil {
		// limits
		if microService.Resources.Limits.CPU == "" {
			microService.Resources.Limits.CPU = "500m"
		}
		if microService.Resources.Limits.Memory == "" {
			microService.Resources.Limits.Memory = "256Mi"
		}

		// requests
		if microService.Resources.Requests.CPU == "" {
			microService.Resources.Requests.CPU = "500m"
		}
		if microService.Resources.Requests.Memory == "" {
			microService.Resources.Requests.Memory = "256Mi"
		}
	}
}
