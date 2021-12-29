package commerce

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/gateway_service.go -package=Mocks github.com/localrivet/k8sly/controllers/commerce Service
// GatewayService interface.
type GatewayService interface {
	Create(cr *cachev1alpha1.Commerce) *corev1.Service
}

// NewGatewayService creates a real implementation of the Service interface.
func NewGatewayService() GatewayService {
	return &gatewayService{}
}

type gatewayService struct{}

// Create returns a new service.
func (s *gatewayService) Create(cr *cachev1alpha1.Commerce) *corev1.Service {
	if cr.Spec.CoreMicroServices.GatewayClient == nil {
		return &corev1.Service{}
	}

	var port int32 = 80
	if cr.Spec.Hosts.Client.Scheme == "https" {
		port = 443
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gateway-client-service",
			Namespace: cr.Spec.TargetNamespace,
			// Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": cr.Spec.CoreMicroServices.GatewayClient.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       cr.Spec.Hosts.Client.Scheme,
					Protocol:   corev1.ProtocolTCP,
					Port:       port,
					TargetPort: intstr.FromInt(int(cr.Spec.CoreMicroServices.GatewayClient.ContainerPort)),
				},
			},
		},
	}
}

func (r *CommerceReconciler) reconcileGatewayService(ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) (ctrl.Result, error) {
	// Define a new Service object
	s := NewGatewayService()
	found := &corev1.Service{}
	wanted := s.Create(cr)

	// Set Deployment instance as the owner and controller of the Service
	// this code creates: cross-namespace owner references are disallowed, owner's namespace default, obj's namespace test
	// if err := controllerutil.SetControllerReference(cr, wanted, r.Scheme); err != nil {
	// 	return ctrl.Result{}, err
	// }

	// Check if this Service already exists
	err := r.Get(ctx, types.NamespacedName{Name: wanted.Name, Namespace: wanted.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", wanted.Namespace, "Service.Name", wanted.Name)
		err = r.Create(ctx, wanted)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Service created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Service already exists
		log.Info("Service already exists", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
	}

	// Service reconcile finished
	return ctrl.Result{}, nil
}
