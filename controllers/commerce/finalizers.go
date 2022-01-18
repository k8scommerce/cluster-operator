package commerce

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// addFinalizer adds a given finalizer to a given CR.
func (r *K8sCommerceReconciler) addFinalizer(ctx context.Context, cr *cachev1alpha1.K8sCommerce, log logr.Logger) error {
	controllerutil.AddFinalizer(cr, commerceFinalizer)

	// Update CR
	if err := r.Update(ctx, cr); err != nil {
		log.Error(err, "Failed to update Deployment with finalizer")
		return err
	}
	return nil
}

// handleFinalizer runs required tasks before deleting the objects owned by the CR.
func (r *K8sCommerceReconciler) handleFinalizer(ctx context.Context, cr *cachev1alpha1.K8sCommerce, log logr.Logger) (err error) {
	if !cr.HasFinalizer(commerceFinalizer) {
		return nil
	}

	// delete the microservices
	{
		// delete the gateway cliet service
		if cr.Spec.CoreMicroServices.GatewayClient != nil {
			if err = r.deleteDeployment(ctx, cr, NewGatewayClient()); err != nil {
				return err
			}
		}

		// delete the gateway admin service
		if cr.Spec.CoreMicroServices.GatewayAdmin != nil {
			if err = r.deleteDeployment(ctx, cr, NewGatewayAdmin()); err != nil {
				return err
			}
		}

		// delete the cart service
		if cr.Spec.CoreMicroServices.Cart != nil {
			if err = r.deleteDeployment(ctx, cr, NewCart()); err != nil {
				return err
			}
		}

		// delete the customer service
		if cr.Spec.CoreMicroServices.Customer != nil {
			if err = r.deleteDeployment(ctx, cr, NewCustomer()); err != nil {
				return err
			}
		}

		// delete the email service
		if cr.Spec.CoreMicroServices.Email != nil {
			if err = r.deleteDeployment(ctx, cr, NewEmail()); err != nil {
				return err
			}
		}

		// delete the inventory service
		if cr.Spec.CoreMicroServices.Inventory != nil {
			if err = r.deleteDeployment(ctx, cr, NewInventory()); err != nil {
				return err
			}
		}

		// delete the othersBought service
		if cr.Spec.CoreMicroServices.OthersBought != nil {
			if err = r.deleteDeployment(ctx, cr, NewOthersBought()); err != nil {
				return err
			}
		}

		// delete the payment service
		if cr.Spec.CoreMicroServices.Payment != nil {
			if err = r.deleteDeployment(ctx, cr, NewPayment()); err != nil {
				return err
			}
		}

		// delete the product service
		if cr.Spec.CoreMicroServices.Product != nil {
			if err = r.deleteDeployment(ctx, cr, NewProduct()); err != nil {
				return err
			}
		}

		// delete the shipping service
		if cr.Spec.CoreMicroServices.Shipping != nil {
			if err = r.deleteDeployment(ctx, cr, NewShipping()); err != nil {
				return err
			}
		}

		// delete the similarProducts service
		if cr.Spec.CoreMicroServices.SimilarProducts != nil {
			if err = r.deleteDeployment(ctx, cr, NewSimilarProducts()); err != nil {
				return err
			}
		}

		// delete the store service
		if cr.Spec.CoreMicroServices.Store != nil {
			if err = r.deleteDeployment(ctx, cr, NewStore()); err != nil {
				return err
			}
		}

		// delete the user service
		if cr.Spec.CoreMicroServices.User != nil {
			if err = r.deleteDeployment(ctx, cr, NewUser()); err != nil {
				return err
			}
		}

		// delete the warehouse service
		if cr.Spec.CoreMicroServices.Warehouse != nil {
			if err = r.deleteDeployment(ctx, cr, NewWarehouse()); err != nil {
				return err
			}
		}
	}

	// delete the gateway client service
	if cr.Spec.CoreMicroServices.GatewayClient != nil {
		if err = r.deleteService(ctx, cr, NewGatewayClientService().Create(cr)); err != nil {
			return err
		}

		// delete the gateway client ingress
		if err = r.deleteIngress(ctx, cr, NewGatewayClientIngress().Create(cr)); err != nil {
			return err
		}
	}

	// delete the gateway admin service
	if cr.Spec.CoreMicroServices.GatewayAdmin != nil {
		if err = r.deleteService(ctx, cr, NewGatewayAdminService().Create(cr)); err != nil {
			return err
		}

		// delete the gateway admin ingress
		if err = r.deleteIngress(ctx, cr, NewGatewayAdminIngress().Create(cr)); err != nil {
			return err
		}
	}

	// delete etcd
	{
		if err = r.deleteService(ctx, cr, NewEtcd().CreateClientService(cr)); err != nil {
			return err
		}

		if err = r.deleteService(ctx, cr, NewEtcd().CreateHeadlessService(cr)); err != nil {
			return err
		}

		var i int32
		for i = 0; i < *cr.Spec.Etcd.Replicas; i++ {
			if err = r.deleteService(ctx, cr, NewEtcd().CreatePodService(cr, i)); err != nil {
				return err
			}

			if err = r.deleteEtcdPod(ctx, cr, i); err != nil {
				return err
			}
		}
	}

	// delete the target namespace
	if err = r.namespaceFinalDelete(ctx, cr); err != nil {
		return err
	}

	// delete the crd
	if err = r.Delete(ctx, cr); err != nil {
		return err
	}

	// remove the finalizer to allow object deletion
	controllerutil.RemoveFinalizer(cr, commerceFinalizer)
	return r.Update(ctx, cr)
}

func (r *K8sCommerceReconciler) deleteDeployment(ctx context.Context, cr *cachev1alpha1.K8sCommerce, s FinalizableDeployment) error {
	dep := s.Create(cr)
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = r.Delete(ctx, dep)
	if err != nil {
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deleted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *K8sCommerceReconciler) deleteEtcdOperator(ctx context.Context, dep *appsv1.Deployment) error {
	// s := NewEtcd()
	// dep := s.CreateOperator()
	depFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = r.Delete(ctx, dep)
	if err != nil {
		r.Log.Error(err, "Failed to delete Deployment")
		return err
	}
	r.Log.Info(fmt.Sprintf("Deployment deleted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *K8sCommerceReconciler) deleteEtcdPod(ctx context.Context, cr *cachev1alpha1.K8sCommerce, id int32) error {
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
	err = r.Delete(ctx, dep)
	if err != nil {
		r.Log.Error(err, "Failed to delete Pod")
		return err
	}
	r.Log.Info(fmt.Sprintf("Pod deleted: %s", "Name: "+dep.Name+" Namespace: "+dep.Namespace))

	return nil
}

func (r *K8sCommerceReconciler) deleteIngress(ctx context.Context, cr *cachev1alpha1.K8sCommerce, ing *networking.Ingress) error {
	ingress := &networking.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: ing.Name, Namespace: ing.Namespace}, ingress)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = r.Delete(ctx, ingress)
	if err != nil {
		r.Log.Error(err, "Failed to delete Ingress")
		return err
	}
	r.Log.Info(fmt.Sprintf("Ingress deleted: %s", "Name: "+ing.Name+" Namespace: "+ing.Namespace))

	return nil
}

func (r *K8sCommerceReconciler) deleteService(ctx context.Context, cr *cachev1alpha1.K8sCommerce, svc *corev1.Service) error {
	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = r.Delete(ctx, service)
	if err != nil {
		r.Log.Error(err, "Failed to delete Service")
		return err
	}
	r.Log.Info(fmt.Sprintf("Service deleted: %s", "Name: "+svc.Name+" Namespace: "+svc.Namespace))

	return nil
}
