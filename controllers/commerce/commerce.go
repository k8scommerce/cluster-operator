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
	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
	"github.com/localrivet/k8sly-operator/controllers/constant"
)

// CommerceReconciler reconciles a Commerce object
type CommerceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

var commerceFinalizer = "cache.commerce.k8sly.com/finalizer"

//+kubebuilder:rbac:groups=cache.commerce.k8sly.com,resources=commerces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.commerce.k8sly.com,resources=commerces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.commerce.k8sly.com,resources=commerces/finalizers,verbs=update

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

	if commerce.IsBeingDeleted() {
		if err := r.handleFinalizer(ctx, commerce, log); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %w", err)
		}
	}

	if !commerce.HasFinalizer(commerceFinalizer) {
		r.Log.Info(fmt.Sprintf("AddFinalizer for %v", req.NamespacedName))
		if err := r.addFinalizer(ctx, commerce, log); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %w", err)
		}
	}

	// Create and/or sync the Secrets from yaml to a Secret
	// if requeue, err := r.secretsSync(ctx, deploymentFound, isNewGeneration); err != nil || requeue {
	// 	return ctrl.Result{Requeue: requeue}, err
	// }

	// Notify START status.
	// if requeue, err := r.notifier(ctx, deploymentFound, models.JobStatusProcessing); err != nil || requeue {
	// 	r.Recorder.Event(deploymentFound, corev1.EventTypeWarning, "Notifier start error", err.Error())
	// 	return ctrl.Result{Requeue: requeue}, err
	// }

	// found := &appsv1.Deployment{}
	// err = r.Get(ctx, types.NamespacedName{Name: commerce.Name, Namespace: commerce.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	// Define a new deployment
	// 	dep := r.deploymentForCommerce(commerce)
	// 	log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 	err = r.Create(ctx, dep)
	// 	if err != nil {
	// 		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	// Deployment created successfully - return and requeue
	// 	return ctrl.Result{Requeue: true}, nil
	// } else if err != nil {
	// 	log.Error(err, "Failed to get Deployment")
	// 	return ctrl.Result{}, err
	// }

	// isNewGeneration := false
	// if commerce.Generation != commerce.Status.ObservedGeneration {
	// 	isNewGeneration = true
	// 	patch := client.MergeFrom(commerce.DeepCopy())
	// 	commerce.Status.ObservedGeneration = commerce.Generation
	// 	if err = r.Status().Patch(ctx, commerce, patch); err != nil {
	// 		r.Log.Error(err, "Failed to update Commerce status during Reconcile")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// // Create and/or sync the Secrets from yaml to a Secret
	// if requeue, err := r.secretsSync(ctx, commerce); err != nil || requeue {
	// 	return ctrl.Result{Requeue: requeue}, err
	// }

	if commerce.Spec.TargetNamespace != "" {
		// Reconcile Target Namespace
		result, err = r.reconcileNamespace(ctx, commerce, log)
		if err != nil {
			return result, err
		}

		// Reconcile Etc
		result, err = r.reconcileEtcd(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.Product != nil {
		// Reconcile Product object
		result, err = r.reconcileProduct(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.Inventory != nil {
		// Reconcile Inventory object
		result, err = r.reconcileInventory(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.OthersBought != nil {
		// Reconcile OthersBought object
		result, err = r.reconcileOthersBought(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.SimilarProducts != nil {
		// Reconcile SimilarProducts object
		result, err = r.reconcileSimilarProducts(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.User != nil {
		// Reconcile User object
		result, err = r.reconcileUser(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.Cart != nil {
		// Reconcile Cart object
		result, err = r.reconcileCart(ctx, commerce, log)
		if err != nil {
			return result, err
		}
	}

	if commerce.Spec.CoreServices.GatewayClient != nil {
		// Reconcile GatwayClient object
		result, err = r.reconcileGatewayClient(ctx, commerce, log)
		if err != nil {
			return result, err
		}

		// Reconcile GatwayService object
		result, err = r.reconcileGatewayService(ctx, commerce, log)
		if err != nil {
			return result, err
		}

		// Reconcile GatwayIngress object
		result, err = r.reconcileGatewayIngress(ctx, commerce, log)
		if err != nil {
			return result, err
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
	var running int32 = 0

	podList, err := r.findComponentsByLabel(cr.Spec.TargetNamespace, client.MatchingLabels{
		"app": "etcd",
	})

	if err != nil {
		r.Log.Error(err, "podlist")
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase == v1.PodRunning {
			running++
			// r.Log.Info("EtcdPod", "Pod is running", running, "pod.Status", pod.Status)
		}
	}

	return running
}
