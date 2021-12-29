package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/gatewayClient.go -package=Mocks github.com/localrivet/k8sly/controllers/commerce mode Deployment
// GatewayClient interface.
type GatewayClient interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

// NewGatewayClient creates a new gatewayClient.
func NewGatewayClient() GatewayClient {
	return &gatewayClient{}
}

type gatewayClient struct{}

// Create Returns a new gatewayClient without replicas configured - replicas will be configured in the sync loop.
func (d *gatewayClient) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.GatewayClient)
	return dep.Create(cr)
}
