/*
Copyright 2020 cnych.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commerce

import (
	"context"

	etcdv1beta2 "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
	"github.com/k8scommerce/cluster-operator/controllers/constant"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	EtcdImage = "quay.io/coreos/etcd-operator:v0.9.4"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/etcd.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Etcd interface.
type Etcd interface {
	CreateCRD() *extapi.CustomResourceDefinition
	CreateOperator() *appsv1.Deployment
	CreateCluster() *etcdv1beta2.EtcdCluster
}

// NewEtcd creates a new etcd.
func NewEtcd() Etcd {
	return &etcd{}
}

type etcd struct{}

// apiVersion: extapi.k8s.io/v1beta1
// kind: CustomResourceDefinition
// metadata:
//  name: etcdclusters.etcd.database.coreos.com
// spec:
//  group: etcd.database.coreos.com
//  names:
//  kind: EtcdCluster
//  listKind: EtcdClusterList
//  plural: etcdclusters
//  shortNames:
//  - etcdclus
//  - etcd
//  singular: etcdcluster
//  scope: Namespaced
//  version: v1beta2
//  versions:
//  - name: v1beta2
//  served: true
//  storage: true
func (d *etcd) CreateCRD() *extapi.CustomResourceDefinition {
	return &extapi.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcdclusters.etcd.database.coreos.com",
			Namespace: constant.ManagerNamespace,
		},
		Spec: extapi.CustomResourceDefinitionSpec{
			Group: "etcd.database.coreos.com",
			Names: extapi.CustomResourceDefinitionNames{
				Kind:     "EtcdCluster",
				ListKind: "EtcdClusterList",
				Plural:   "etcdclusters",
				ShortNames: []string{
					"etcdclus",
					"etcd",
				},
				Singular: "etcdcluster",
			},
			Scope: extapi.NamespaceScoped,
			Versions: []extapi.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
		},
	}
}

// apiVersion: extensions/v1beta1
// kind: Deployment
// metadata:
//   name: etcd-operator
// spec:
//   replicas: 1
//   template:
//     metadata:
//       labels:
//         name: etcd-operator
//     spec:
//       containers:
//       - name: etcd-operator
//         image: quay.io/coreos/etcd-operator:v0.9.4
//         command:
//         - etcd-operator
//         # Uncomment to act for resources in all namespaces. More information in doc/user/clusterwide.md
//         #- -cluster-wide
//         env:
//         - name: MY_POD_NAMESPACE
//           valueFrom:
//             fieldRef:
//               fieldPath: metadata.namespace
//         - name: MY_POD_NAME
//           valueFrom:
//             fieldRef:
//               fieldPath: metadata.name
func (d *etcd) CreateOperator() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-operator",
			Namespace: constant.ManagerNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: NewInt32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "etcd-operator",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "etcd-operator",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(5),
					Containers: []corev1.Container{
						{
							Name:  "etcd-operator",
							Image: EtcdImage,
							Command: []string{
								"etcd-operator",
							},
							Env: []corev1.EnvVar{
								{
									Name: "MY_POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "MY_POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// apiVersion: "etcd.database.coreos.com/v1beta2"
// kind: "EtcdCluster"
// metadata:
//   name: "example-etcd-cluster"
//   ## Adding this annotation make this cluster managed by clusterwide operators
//   ## namespaced operators ignore it
//   # annotations:
//   #   etcd.database.coreos.com/scope: clusterwide
// spec:
//   size: 3
//   version: "3.2.13"
func (d *etcd) CreateCluster() *etcdv1beta2.EtcdCluster {
	return &etcdv1beta2.EtcdCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd",
			Namespace: constant.TargetNamespace,
		},
		Spec: etcdv1beta2.ClusterSpec{
			Size:    3,
			Version: "3.2.13",
		},
	}
}

//
// Reconcile Functions.
//

func reconcileEtcdCrd(r *K8sCommerceReconciler, etcd Etcd, ctx context.Context, cr *cachev1alpha1.K8sCommerce, log logr.Logger) (ctrl.Result, error) {
	foundEtcdCRD := &extapi.CustomResourceDefinition{}
	wantedEtcdCRD := etcd.CreateCRD()
	err := r.Get(ctx, types.NamespacedName{Name: wantedEtcdCRD.Name, Namespace: wantedEtcdCRD.Namespace}, foundEtcdCRD)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new EtcdOperator", "Etcd.Namespace", wantedEtcdCRD.Namespace, "EtcdOperator.Name", wantedEtcdCRD.Name)
		err = r.Create(ctx, wantedEtcdCRD)
		if err != nil {
			log.Error(err, "Failed to create new EtcdOperator", "Etcd.Namespace", wantedEtcdCRD.Namespace, "EtcdOperator.Name", wantedEtcdCRD.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod EtcdOperator")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileEtcdOperator(r *K8sCommerceReconciler, etcd Etcd, ctx context.Context, cr *cachev1alpha1.K8sCommerce, log logr.Logger) (ctrl.Result, error) {
	foundEtcdOperator := &appsv1.Deployment{}
	wantedEtcdOperator := etcd.CreateOperator()
	err := r.Get(ctx, types.NamespacedName{Name: wantedEtcdOperator.Name, Namespace: wantedEtcdOperator.Namespace}, foundEtcdOperator)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new EtcdOperator", "Etcd.Namespace", wantedEtcdOperator.Namespace, "EtcdOperator.Name", wantedEtcdOperator.Name)
		err = r.Create(ctx, wantedEtcdOperator)
		if err != nil {
			log.Error(err, "Failed to create new EtcdOperator", "Etcd.Namespace", wantedEtcdOperator.Namespace, "EtcdOperator.Name", wantedEtcdOperator.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod EtcdOperator")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileEtcCluster(r *K8sCommerceReconciler, etcd Etcd, ctx context.Context, cr *cachev1alpha1.K8sCommerce, log logr.Logger) (ctrl.Result, error) {
	foundEtcdCluster := &etcdv1beta2.EtcdCluster{}
	wantedEtcdCluster := etcd.CreateCluster()
	err := r.Get(ctx, types.NamespacedName{Name: wantedEtcdCluster.Name, Namespace: wantedEtcdCluster.Namespace}, foundEtcdCluster)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new EtcdCluster", "Etcd.Namespace", wantedEtcdCluster.Namespace, "EtcdCluster.Name", wantedEtcdCluster.Name)
		err = r.Create(ctx, wantedEtcdCluster)
		if err != nil {
			log.Error(err, "Failed to create new EtcdCluster", "Etcd.Namespace", wantedEtcdCluster.Namespace, "EtcdCluster.Name", wantedEtcdCluster.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod EtcdCluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *K8sCommerceReconciler) reconcileEtcd(ctx context.Context, cr *cachev1alpha1.K8sCommerce, log logr.Logger) (ctrl.Result, error) {
	// Define a new Deployment object
	etcd := NewEtcd()
	result, err := reconcileEtcdCrd(r, etcd, ctx, cr, log)
	if err != nil {
		return result, err
	}

	result, err = reconcileEtcdOperator(r, etcd, ctx, cr, log)
	if err != nil {
		return result, err
	}

	result, err = reconcileEtcCluster(r, etcd, ctx, cr, log)
	if err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}
