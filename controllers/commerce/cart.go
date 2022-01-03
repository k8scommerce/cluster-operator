package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/cart.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Cart interface.
type Cart interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewCart creates a new cart.
func NewCart() Cart {
	return &cart{}
}

type cart struct{}

// Create Returns a new cart without replicas configured - replicas will be configured in the sync loop.
func (d *cart) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Cart)
	return dep.Create(cr)
}
