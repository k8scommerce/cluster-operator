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

	// delete etcd
	{
		var i int32
		for i = 0; i < *cr.Spec.Etcd.Replicas; i++ {
			if err = r.deleteEtcdPod(ctx, cr, i); err != nil {
				return err
			}
			if err = r.deleteService(ctx, cr, NewEtcd().CreatePodService(cr, i)); err != nil {
				return err
			}
		}
		if err = r.deleteService(ctx, cr, NewEtcd().CreateClientService(cr)); err != nil {
			return err
		}
	}

	// delete the microservices
	{
		// delete the gateway cliet service
		if err = r.deleteDeployment(ctx, cr, NewGatewayClient()); err != nil {
			return err
		}

		// delete the cart service
		if err = r.deleteDeployment(ctx, cr, NewCart()); err != nil {
			return err
		}

		// delete the inventory service
		if err = r.deleteDeployment(ctx, cr, NewInventory()); err != nil {
			return err
		}

		// delete the othersBought service
		if err = r.deleteDeployment(ctx, cr, NewOthersBought()); err != nil {
			return err
		}

		// delete the similarProducts service
		if err = r.deleteDeployment(ctx, cr, NewSimilarProducts()); err != nil {
			return err
		}

		// delete the user service
		if err = r.deleteDeployment(ctx, cr, NewUser()); err != nil {
			return err
		}

		// delete the product service
		if err = r.deleteDeployment(ctx, cr, NewProduct()); err != nil {
			return err
		}
	}

	// delete the gateway service
	if err = r.deleteService(ctx, cr, NewGatewayService().Create(cr)); err != nil {
		return err
	}

	// delete the gateway ingress
	if err = r.deleteIngress(ctx, cr, NewGatewayIngress().Create(cr)); err != nil {
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

func (r *CommerceReconciler) deleteDeployment(ctx context.Context, cr *cachev1alpha1.Commerce, s FinalizableDeployment) error {
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deleted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteEtcdPod(ctx context.Context, cr *cachev1alpha1.Commerce, id int32) error {
	s := NewEtcd()
	dep := s.CreatePod(cr, id)
	depFound := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = r.Client.Delete(ctx, dep)
	if err != nil {
		r.Log.Error(err, "Failed to delete Pod")
		return err
	}
	r.Log.Info(fmt.Sprintf("Pod deleted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *CommerceReconciler) deleteIngress(ctx context.Context, cr *cachev1alpha1.Commerce, ing *networking.Ingress) error {
	ingress := &networking.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: ing.Name, Namespace: ing.Namespace}, ingress)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
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

func (r *CommerceReconciler) deleteService(ctx context.Context, cr *cachev1alpha1.Commerce, svc *corev1.Service) error {
	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
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
