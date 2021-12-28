/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// ConfigMapRef is the name of the environment variable configMap created by the operator-environment.
const ConfigMapRef = "k8sly-config"

// SecretRef is the name of the secret created by the operator-environment.
const SecretRef = "k8sly-secret"

type CommerceHost struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int32  `json:"port,omitempty"`
}

type CommerceHosts struct {
	Client CommerceHost `json:"client,omitempty"`
	Admin  CommerceHost `json:"admin,omitempty"`
}

type CommerceResources struct {
}

type CommerceService struct {
	Name            string            `json:"name,omitempty"`
	Image           string            `json:"image,omitempty"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	ContainerPort   int32             `json:"port,omitempty"`
	CPU             string            `json:"cpu,omitempty"`
	Memory          string            `json:"memory,omitempty"`
	EnvironmentVars map[string]string `json:"vars,omitempty"`

	//+kubebuilder:validation:Minimum=2
	Replicas int32 `json:"replicas,omitempty"`
}

type CommerceCoreServices struct {
	GatewayClient   *CommerceService `json:"gatewayClient,omitempty"`
	User            *CommerceService `json:"user,omitempty"`
	Product         *CommerceService `json:"product,omitempty"`
	Cart            *CommerceService `json:"cart,omitempty"`
	Inventory       *CommerceService `json:"inventory,omitempty"`
	OthersBought    *CommerceService `json:"othersBought,omitempty"`
	SimilarProducts *CommerceService `json:"similarProducts,omitempty"`
}

type CommerceDatabaseConnection struct {
	SecretName string `json:"secretName,omitempty"`
}

type CommerceEtcd struct {
	Replicas *int32 `json:"replicas,omitempty"`
}

// CommerceSpec defines the desired state of Commerce
type CommerceSpec struct {
	TargetNamespace string                      `json:"targetNamespace"`
	CoreServices    CommerceCoreServices        `json:"coreServices"`
	AddOnServices   []CommerceService           `json:"addOnServices,omitempty"`
	Hosts           CommerceHosts               `json:"hosts"`
	Database        *CommerceDatabaseConnection `json:"database"`
	Etcd            CommerceEtcd                `json:"etcd"`
}

// CommerceStatus defines the observed state of Commerce
type CommerceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Commerce is the Schema for the commerces API
type Commerce struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CommerceSpec   `json:"spec,omitempty"`
	Status CommerceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CommerceList contains a list of Commerce
type CommerceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Commerce `json:"items"`
}

// IsBeingDeleted returns true if a deletion timestamp is set.
func (in *Commerce) IsBeingDeleted() bool {
	return !in.ObjectMeta.GetDeletionTimestamp().IsZero()
}

// HasFinalizer returns true if a deletion timestamp is set.
func (in *Commerce) HasFinalizer(finalizerName string) bool {
	return containsString(in.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer.
func (in *Commerce) AddFinalizer(finalizerName string) {
	in.ObjectMeta.Finalizers = append(in.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer.
func (in *Commerce) RemoveFinalizer(finalizerName string) {
	in.ObjectMeta.Finalizers = removeString(in.ObjectMeta.Finalizers, finalizerName)
}

func init() {
	SchemeBuilder.Register(&Commerce{}, &CommerceList{})
}
