package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/gatewayadmin.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// GatewayAdmin interface.
type GatewayAdmin interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewGatewayAdmin creates a new gatewayAdmin.
func NewGatewayAdmin() GatewayAdmin {
	return &gatewayAdmin{}
}

type gatewayAdmin struct{}

// Create Returns a new gatewayAdmin without replicas configured - replicas will be configured in the sync loop.
func (d *gatewayAdmin) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.GatewayAdmin)
	return dep.Create(cr)
}
