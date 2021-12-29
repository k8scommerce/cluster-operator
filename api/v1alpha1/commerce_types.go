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

type Probe struct {
	// +kubebuilder:default:=8080
	Port int `json:"port,omitempty"`
	// +kubebuilder:default:=5
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// +kubebuilder:default:=10
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
}
type Host struct {
	Hostname string `json:"hostname"`
	Scheme   string `json:"scheme"`
}

type Hosts struct {
	Client Host `json:"client"`
	Admin  Host `json:"admin"`
}
type ResourceRequests struct {
	// +kubebuilder:default:="500m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:default:="256Mi"
	Memory string `json:"memory,omitempty"`
}

type ResourceLimits struct {
	// +kubebuilder:default:="500m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:default:="256Mi"
	Memory string `json:"memory,omitempty"`
}

type Resources struct {
	// +optional
	Requests ResourceRequests `json:"requests,omitempty"`
	// +optional
	Limits ResourceLimits `json:"limits,omitempty"`
}
type Lifecycle struct {
	PreStopCommand []string `json:"command,omitempty"`
}

type MicroService struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	// +kubebuilder:validation:Required
	ContainerPort int32 `json:"port,omitempty"`
	// +optional
	EnvironmentVars map[string]string `json:"vars,omitempty"`
	// +optional
	Lifecycle Lifecycle `json:"lifecycle,omitempty"`
	// +optional
	Resources Resources `json:"resources,omitempty"`
	// +optional
	ReadinessProbe Probe `json:"readinessProbe,omitempty"`
	// +optional
	LivenessProbe Probe `json:"livenessProbe,omitempty"`

	//+kubebuilder:validation:Minimum=2
	Replicas int32 `json:"replicas,omitempty"`
}

type CoreMicroServices struct {
	// +kubebuilder:validation:Required
	GatewayClient *MicroService `json:"gatewayClient,omitempty"`
	// +kubebuilder:validation:Required
	User *MicroService `json:"user,omitempty"`
	// +kubebuilder:validation:Required
	Product *MicroService `json:"product,omitempty"`
	// +kubebuilder:validation:Required
	Cart *MicroService `json:"cart,omitempty"`
	// +kubebuilder:validation:Required
	Inventory *MicroService `json:"inventory,omitempty"`
	// +kubebuilder:validation:Required
	OthersBought *MicroService `json:"othersBought,omitempty"`
	// +kubebuilder:validation:Required
	SimilarProducts *MicroService `json:"similarProducts,omitempty"`
}

type Database struct {
	SecretName string `json:"secretName,omitempty"`
}

type Etcd struct {
	Replicas *int32 `json:"replicas,omitempty"`
}

// CommerceSpec defines the desired state of Commerce
type CommerceSpec struct {
	TargetNamespace   string            `json:"targetNamespace"`
	CorsOrigins       []string          `json:"corsOrigins"`
	Hosts             Hosts             `json:"hosts"`
	CoreMicroServices CoreMicroServices `json:"coreMicroServices"`
	AddOnServices     []MicroService    `json:"addOnMicroServices,omitempty"`
	Database          *Database         `json:"database,omitempty"`
	Etcd              Etcd              `json:"etcd"`
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
