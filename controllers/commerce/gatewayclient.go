package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/gatewayclient.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// GatewayClient interface.
type GatewayClient interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

// NewGatewayClient creates a new gatewayClient.
func NewGatewayClient() GatewayClient {
	return &gatewayClient{}
}

type gatewayClient struct{}

// Create Returns a new gatewayClient without replicas configured - replicas will be configured in the sync loop.
func (d *gatewayClient) Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.GatewayClient)
	return dep.Create(cr)
}
