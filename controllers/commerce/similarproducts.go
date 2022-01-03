package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/similarProducts.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// SimilarProducts interface.
type SimilarProducts interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewSimilarProducts creates a new similarProducts.
func NewSimilarProducts() SimilarProducts {
	return &similarProducts{}
}

type similarProducts struct{}

// Create Returns a new similarProducts without replicas configured - replicas will be configured in the sync loop.
func (d *similarProducts) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.SimilarProducts)
	return dep.Create(cr)
}
