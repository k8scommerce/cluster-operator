package commerce

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

// Create namespaceCreate. If return true, requeue.
func (r *CommerceReconciler) reconcileNamespace(ctx context.Context, commerce *cachev1alpha1.Commerce, log logr.Logger) (ctrl.Result, error) {

	r.Log.Info(fmt.Sprintf("Starting namespaceCreate %s", commerce.GetName()))

	ns := &corev1.Namespace{}
	ns.Name = commerce.Spec.TargetNamespace

	err := r.Client.Get(ctx, types.NamespacedName{Name: ns.Name}, ns)
	if err != nil && errors.IsNotFound(err) {
		r.Log.Info(fmt.Sprintf("Namespace not found, attempting to create %s", ns.Name))
		err = r.Client.Create(ctx, ns)
		if err != nil {
			r.Log.Error(err, "Failed to create namespaceCreate")
			// r.Recorder.Event(commerce, corev1.EventTypeWarning, "Created Namespace Failed", fmt.Sprintf("Namespace: %s", ns.Name))
			return ctrl.Result{}, err
		}
		r.Log.Info(fmt.Sprintf("Namespace created: %s", ns.Name))
		// r.Recorder.Event(commerce, corev1.EventTypeNormal, "Namespace Created", fmt.Sprintf("Namespace: %s", ns.Name))

		patch := client.MergeFrom(commerce.DeepCopy())
		// commerce.ObjectMeta.Namespace = commerce.ObjectMeta.Name
		if err = r.Status().Patch(ctx, commerce, patch); err != nil {
			// r.Recorder.Event(commerce, corev1.EventTypeWarning, "Failed Status Update", fmt.Sprintf("Error: %s: Existing: %s, Requested:%s", err.Error(), commerce.Status.Name, commerce.ObjectMeta.Name))

			r.Log.Error(err, "Failed to update Commerce status during namespaceCreate create")
			return ctrl.Result{}, err
		}
		// r.Recorder.Event(commerce, corev1.EventTypeNormal, "Status Updated", fmt.Sprintf("Existing: %s, Requested:%s", commerce.Status.Name, commerce.ObjectMeta.Name))

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// Create namespaceFinalDelete. If return true, requeue.
func (r *CommerceReconciler) namespaceFinalDelete(ctx context.Context, commerce *cachev1alpha1.Commerce) error {

	if commerce.ObjectMeta.Namespace != "" {
		ns := &corev1.Namespace{}
		ns.Name = commerce.Spec.TargetNamespace
		if err := r.Client.Get(ctx, types.NamespacedName{Name: ns.Name}, ns); err == nil {

			// r.Recorder.Event(commerce, corev1.EventTypeNormal, "Deleting Namespace", fmt.Sprintf("Namespace: %s", ns.Name))
			err = r.Client.Delete(ctx, ns)
			if err != nil {
				// r.Recorder.Event(commerce, corev1.EventTypeWarning, "Delete Namespace Failed", fmt.Sprintf("Namespace: %s", ns.Name))
				r.Log.Error(err, "Failed to delete namespaceFinalDelete")
				return err
			}
			r.Log.Info(fmt.Sprintf("Namespace deleted: %s", ns.Name))
		}

	}

	return nil
}
