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
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
	"github.com/k8scommerce/cluster-operator/controllers/constant"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	EtcdImage = "quay.io/coreos/etcd:latest"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/etcd.go -package=Mocks github.com/k8scommerce/k8scommerce/controllers/commerce mode Deployment
// Etcd interface.
type Etcd interface {
	CreateClientService(cr *cachev1alpha1.Commerce) *corev1.Service
	CreatePodService(cr *cachev1alpha1.Commerce, id int32) *corev1.Service
	CreatePod(cr *cachev1alpha1.Commerce, id int32) *corev1.Pod
	// HasVersionMismatch(current *appsv1.Deployment, desired *appsv1.Deployment) bool
	// IsReady(etcd *appsv1.Deployment) bool
	// isEnvHashCurrent(etcd *appsv1.Deployment, annotationKey string, hash string) bool
}

// NewEtcd creates a new etcd.
func NewEtcd() Etcd {
	return &etcd{}
}

type etcd struct{}

// apiVersion: v1
// kind: Service
// metadata:
//   name: etcd-client
//   namespace: k8scommerce-system
// spec:
//   ports:
//     - name: etcd-client-port
//       port: 2379
//       protocol: TCP
//       targetPort: 2379
//   selector:
//     app: etcd
func (d *etcd) CreateClientService(cr *cachev1alpha1.Commerce) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-client",
			Namespace: constant.TargetNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "etcd-client-port",
					Port:       2379,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(2379),
				},
			},
			Selector: map[string]string{
				"app": "etcd",
			},
		},
	}

	return svc
}

// apiVersion: v1
// kind: Service
// metadata:
//   labels:
//     etcd_node: etcd0
//   name: etcd0
//   namespace: k8scommerce-system
// spec:
//   ports:
//     - name: client
//       port: 2379
//       protocol: TCP
//       targetPort: 2379
//     - name: server
//       port: 2380
//       protocol: TCP
//       targetPort: 2380
//   selector:
//     etcd_node: etcd0
func (d *etcd) CreatePodService(cr *cachev1alpha1.Commerce, id int32) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"etcd_node": fmt.Sprintf("etcd%d", id),
			},
			Name:      fmt.Sprintf("etcd%d", id),
			Namespace: constant.TargetNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "client",
					Port:       2379,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(2379),
				}, {
					Name:       "server",
					Port:       2380,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(2380),
				},
			},
			Selector: map[string]string{
				"etcd_node": fmt.Sprintf("etcd%d", id),
			},
		},
	}

	return svc
}

// apiVersion: v1
// kind: Pod
// metadata:
//   labels:
//     app: etcd
//     etcd_node: etcd0
//   name: etcd0
//   namespace: k8scommerce-system
// spec:
//   containers:
//     - command:
//         - /usr/local/bin/etcd
//         - --name
//         - etcd0
//         - --initial-advertise-peer-urls
//         - http://etcd0:2380
//         - --listen-peer-urls
//         - http://0.0.0.0:2380
//         - --listen-client-urls
//         - http://0.0.0.0:2379
//         - --advertise-client-urls
//         - http://etcd0:2379
//         - --initial-cluster
//         - etcd0=http://etcd0:2380,etcd1=http://etcd1:2380,etcd2=http://etcd2:2380
//         - --initial-cluster-state
//         - new
//       image: quay.io/coreos/etcd:latest
//       name: etcd0
//       ports:
//         - containerPort: 2379
//           name: client
//           protocol: TCP
//         - containerPort: 2380
//           name: server
//           protocol: TCP
//   restartPolicy: Always
func (d *etcd) CreatePod(cr *cachev1alpha1.Commerce, id int32) *corev1.Pod {
	var etcdEndpoints []string
	var x int32
	for x = 0; x < *cr.Spec.Etcd.Replicas; x++ {
		// "etcd0=http://etcd0:2380,etcd1=http://etcd1:2380,etcd2=http://etcd2:2380",
		etcdEndpoints = append(etcdEndpoints, fmt.Sprintf("etcd%d=http://etcd%d:2380", x, x))
	}

	etcdEndpoint := strings.Join(etcdEndpoints, ",")

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":       "etcd",
				"etcd_node": fmt.Sprintf("etcd%d", id),
			},
			Name:      fmt.Sprintf("etcd%d", id),
			Namespace: constant.TargetNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: EtcdImage,
					Name:  fmt.Sprintf("etcd%d", id),
					Ports: []corev1.ContainerPort{
						{
							Name:          "client",
							ContainerPort: 2379,
							Protocol:      corev1.ProtocolTCP,
						}, {
							Name:          "server",
							ContainerPort: 2380,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Command: []string{
						"/usr/local/bin/etcd",
						"--name",
						fmt.Sprintf("etcd%d", id),
						"--initial-advertise-peer-urls",
						fmt.Sprintf("http://etcd%d:2380", id),
						"--listen-peer-urls",
						"http://0.0.0.0:2380",
						"--listen-client-urls",
						"http://0.0.0.0:2379",
						"--advertise-client-urls",
						fmt.Sprintf("http://etcd%d:2379", id),
						"--initial-cluster",
						etcdEndpoint,
						"--initial-cluster-state",
						"new",
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}

	return pod
}

//
// Reconcile Functions.
//

func reconcileEtcdClientService(r *CommerceReconciler, etcd Etcd, ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) (ctrl.Result, error) {
	foundClientService := &corev1.Service{}
	wantedClientService := etcd.CreateClientService(cr)
	err := r.Get(ctx, types.NamespacedName{Name: wantedClientService.Name, Namespace: wantedClientService.Namespace}, foundClientService)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Etcd.Namespace", wantedClientService.Namespace, "Service.Name", wantedClientService.Name)
		err = r.Create(ctx, wantedClientService)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Etcd.Namespace", wantedClientService.Namespace, "Service.Name", wantedClientService.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Client Service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileEtcdPodServices(r *CommerceReconciler, etcd Etcd, ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger, id int32) (ctrl.Result, error) {
	foundClientService := &corev1.Service{}
	wantedClientService := etcd.CreatePodService(cr, id)
	err := r.Get(ctx, types.NamespacedName{Name: wantedClientService.Name, Namespace: wantedClientService.Namespace}, foundClientService)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Etcd.Namespace", wantedClientService.Namespace, "Service.Name", wantedClientService.Name)
		err = r.Create(ctx, wantedClientService)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Etcd.Namespace", wantedClientService.Namespace, "Service.Name", wantedClientService.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod Service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileEtcdPods(r *CommerceReconciler, etcd Etcd, ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger, id int32) (ctrl.Result, error) {
	foundClientService := &corev1.Pod{}
	wantedClientService := etcd.CreatePod(cr, id)
	err := r.Get(ctx, types.NamespacedName{Name: wantedClientService.Name, Namespace: wantedClientService.Namespace}, foundClientService)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Etcd.Namespace", wantedClientService.Namespace, "Service.Name", wantedClientService.Name)
		err = r.Create(ctx, wantedClientService)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Etcd.Namespace", wantedClientService.Namespace, "Service.Name", wantedClientService.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod Service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CommerceReconciler) reconcileEtcd(ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) (ctrl.Result, error) {
	// Define a new Deployment object
	etcd := NewEtcd()
	result, err := reconcileEtcdClientService(r, etcd, ctx, cr, log)
	if err != nil {
		return result, err
	}

	var id int32
	for id = 0; id < *cr.Spec.Etcd.Replicas; id++ {
		result, err := reconcileEtcdPodServices(r, etcd, ctx, cr, log, id)
		if err != nil {
			return result, err
		}

		result, err = reconcileEtcdPods(r, etcd, ctx, cr, log, id)
		if err != nil {
			return result, err
		}
	}

	return ctrl.Result{}, nil
}
