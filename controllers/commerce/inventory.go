package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/inventory.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Inventory interface.
type Inventory interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

// NewInventory creates a new inventory.
func NewInventory() Inventory {
	return &inventory{}
}

type inventory struct{}

// Create Returns a new inventory without replicas configured - replicas will be configured in the sync loop.
func (d *inventory) Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Inventory)
	return dep.Create(cr)
}
