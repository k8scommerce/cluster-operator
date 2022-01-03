package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/product.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Product interface.
type Product interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewProduct creates a new product.
func NewProduct() Product {
	return &product{}
}

type product struct{}

// Create Returns a new product without replicas configured - replicas will be configured in the sync loop.
func (d *product) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Product)
	return dep.Create(cr)
}
