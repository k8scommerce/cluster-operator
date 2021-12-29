package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/othersBought.go -package=Mocks github.com/localrivet/k8sly/controllers/commerce mode Deployment
// OthersBought interface.
type OthersBought interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewOthersBought creates a new othersBought.
func NewOthersBought() OthersBought {
	return &othersBought{}
}

type othersBought struct{}

// Create Returns a new othersBought without replicas configured - replicas will be configured in the sync loop.
func (d *othersBought) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.OthersBought)
	return dep.Create(cr)
}
