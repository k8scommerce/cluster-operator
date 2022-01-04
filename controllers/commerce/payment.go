package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/payment.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Payment interface.
type Payment interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

// NewPayment creates a new payment.
func NewPayment() Payment {
	return &payment{}
}

type payment struct{}

// Create Returns a new payment without replicas configured - replicas will be configured in the sync loop.
func (d *payment) Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Payment)
	return dep.Create(cr)
}
