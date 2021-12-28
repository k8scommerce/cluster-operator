package commerce

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

// addFinalizer adds a given finalizer to a given CR.
func (r *CommerceReconciler) addFinalizer(ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) error {
	controllerutil.AddFinalizer(cr, commerceFinalizer)

	// Update CR
	if err := r.Update(ctx, cr); err != nil {
		log.Error(err, "Failed to update Deployment with finalizer")
		return err
	}
	return nil
}

// handleFinalizer runs required tasks before deleting the objects owned by the CR.
func (r *CommerceReconciler) handleFinalizer(ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) (err error) {
	if !cr.HasFinalizer(commerceFinalizer) {
		return nil
	}

	// delete the gateway cliet service
	if err = r.deleteGatewayClient(ctx, cr); err != nil {
		return err
	}

	// delete the cart service
	if err = r.deleteCart(ctx, cr); err != nil {
		return err
	}

	// delete the inventory service
	if err = r.deleteInventory(ctx, cr); err != nil {
		return err
	}

	// delete the othersBought service
	if err = r.deleteOthersBought(ctx, cr); err != nil {
		return err
	}

	// delete the similarProducts service
	if err = r.deleteSimilarProducts(ctx, cr); err != nil {
		return err
	}

	// delete the user service
	if err = r.deleteUser(ctx, cr); err != nil {
		return err
	}

	// delete the product service
	if err = r.deleteProduct(ctx, cr); err != nil {
		return err
	}

	// delete the service
	if err = r.deleteGatewayService(ctx, cr); err != nil {
		return err
	}

	// delete the ingress
	if err = r.deleteGatewayIngress(ctx, cr); err != nil {
		return err
	}

	// delete the target namespace
	if err = r.namespaceFinalDelete(ctx, cr); err != nil {
		return err
	}

	// delete the crd
	if err = r.Client.Delete(ctx, cr); err != nil {
		return err
	}

	// remove the finalizer to allow object deletion
	controllerutil.RemoveFinalizer(cr, commerceFinalizer)
	return r.Update(ctx, cr)
}

func (r *CommerceReconciler) deleteGatewayIngress(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	i := NewGatewayIngress()
	ing := i.Create(cr)
	ingress := &networking.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: ing.Name, Namespace: ing.Namespace}, ingress)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, ingress)
	if err != nil {
		r.Log.Error(err, "Failed to delete Ingress")
		return err
	}
	r.Log.Info(fmt.Sprintf("Ingress deleted: %s", "Name: "+ing.Name+" Namespace: "+ing.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteGatewayService(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewGatewayService()
	svc := s.Create(cr)
	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, service)
	if err != nil {
		r.Log.Error(err, "Failed to delete Service")
		return err
	}
	r.Log.Info(fmt.Sprintf("Service deleted: %s", "Name: "+svc.Name+" Namespace: "+svc.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteProduct(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewProduct()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteCart(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewCart()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteInventory(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewInventory()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteOthersBought(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewOthersBought()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteSimilarProducts(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewSimilarProducts()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteUser(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewUser()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteGatewayClient(ctx context.Context, cr *cachev1alpha1.Commerce) error {
	s := NewGatewayClient()
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// not found, so we are good to go
		return nil
	}
	if err != nil {
		// another type of error, return it
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		// r.Recorder.Event(cr, corev1.EventTypeWarning, "Delete Deployment Failed", "Name: "+dep.Name+" Namespace: "+dep.Namespace)
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deeted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}
