package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/shipping.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Shipping interface.
type Shipping interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

// NewShipping creates a new shipping.
func NewShipping() Shipping {
	return &shipping{}
}

type shipping struct{}

// Create Returns a new shipping without replicas configured - replicas will be configured in the sync loop.
func (d *shipping) Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Shipping)
	return dep.Create(cr)
}
