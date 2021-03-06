package commerce

import (
	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

//
// reconcilers
//
type ReconcilableDeployment interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

//
// finalizers
//

type FinalizableDeployment interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}

type FinalizableService interface {
	Create(cr *cachev1alpha1.K8sCommerce) *appsv1.Deployment
}
