package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/warehouse.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Warehouse interface.
type Warehouse interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewWarehouse creates a new warehouse.
func NewWarehouse() Warehouse {
	return &warehouse{}
}

type warehouse struct{}

// Create Returns a new warehouse without replicas configured - replicas will be configured in the sync loop.
func (d *warehouse) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Warehouse)
	return dep.Create(cr)
}
