package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/store.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Store interface.
type Store interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewStore creates a new store.
func NewStore() Store {
	return &store{}
}

type store struct{}

// Create Returns a new store without replicas configured - replicas will be configured in the sync loop.
func (d *store) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Store)
	return dep.Create(cr)
}
