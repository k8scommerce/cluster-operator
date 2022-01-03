package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/user.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// User interface.
type User interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewUser creates a new user.
func NewUser() User {
	return &user{}
}

type user struct{}

// Create Returns a new user without replicas configured - replicas will be configured in the sync loop.
func (d *user) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.User)
	return dep.Create(cr)
}
