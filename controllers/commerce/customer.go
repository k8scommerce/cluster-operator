package commerce

import (
	appsv1 "k8s.io/api/apps/v1"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/customer.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Customer interface.
type Customer interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

// NewCustomer creates a new customer.
func NewCustomer() Customer {
	return &customer{}
}

type customer struct{}

// Create Returns a new customer without replicas configured - replicas will be configured in the sync loop.
func (d *customer) Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment {
	dep := NewMicroserviceDeployment(cr.Spec.CoreMicroServices.Customer)
	return dep.Create(cr)
}
