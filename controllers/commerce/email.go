package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/email.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Email interface.
type Email interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewEmail creates a new email.
func NewEmail() Email {
	return &email{}
}

type email struct{}

// Create Returns a new email without replicas configured - replicas will be configured in the sync loop.
func (d *email) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Email)
	return dep.Create(cr)
}
