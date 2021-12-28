package commerce

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

//go:generate mockgen -destination ../internal/controllers/commerce/mocks/othersbought.go -package=Mocks github.com/localrivet/k8sly/controllers/commerce mode Deployment
// OthersBought interface.
type OthersBought interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
	HasVersionMismatch(current *appsv1.Deployment, desired *appsv1.Deployment) bool
	IsReady(othersbought *appsv1.Deployment) bool
	isEnvHashCurrent(othersbought *appsv1.Deployment, annotationKey string, hash string) bool
}

// NewOthersBought creates a new othersbought.
func NewOthersBought() OthersBought {
	return &othersbought{}
}

type othersbought struct{}

// Create Returns a new othersbought without replicas configured - replicas will be configured in the sync loop.
func (d *othersbought) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	if cr.Spec.CoreServices.OthersBought == nil {
		return &appsv1.Deployment{}
	}

	annotations := map[string]string{
		"operator-sdk/primary-resource":      fmt.Sprintf("%s/%s", cr.ObjectMeta.Namespace, cr.ObjectMeta.Name),
		"operator-sdk/primary-resource-type": "Commerce.apps",
	}

	var volumeMounts []corev1.VolumeMount
	var volumes []corev1.Volume

	// set the default port
	if cr.Spec.CoreServices.OthersBought != nil && cr.Spec.CoreServices.OthersBought.ContainerPort == 0 {
		cr.Spec.CoreServices.OthersBought.ContainerPort = 8080
	}

	// create the container
	container := corev1.Container{
		Image: CleanContainerImage(cr.Spec.CoreServices.OthersBought.Image),
		Name:  cr.Spec.CoreServices.OthersBought.Name,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: cr.Spec.CoreServices.OthersBought.ContainerPort,
				Name:          "http",
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.CoreServices.OthersBought.CPU),
				corev1.ResourceMemory: resource.MustParse(cr.Spec.CoreServices.OthersBought.Memory),
			},
		},
		VolumeMounts: volumeMounts,
	}

	// does a command exist?
	if len(cr.Spec.CoreServices.OthersBought.Command) > 0 {
		container.Command = cr.Spec.CoreServices.OthersBought.Command
	}

	if len(cr.Spec.CoreServices.OthersBought.Args) > 0 {
		container.Args = cr.Spec.CoreServices.OthersBought.Args
	}

	// check for health path
	// if cr.Spec.HealthPath != "" {
	// 	probe := &corev1.Probe{
	// 		InitialDelaySeconds: 2,
	// 		PeriodSeconds:       1,
	// 		SuccessThreshold:    1,
	// 		TimeoutSeconds:      2,
	// 		FailureThreshold:    6,
	// 		Handler: corev1.Handler{
	// 			HTTPGet: &corev1.HTTPGetAction{
	// 				Port: intstr.FromString("http"),
	// 				Path: cr.Spec.HealthPath,
	// 			},
	// 		},
	// 	}

	// 	container.LivenessProbe = probe
	// 	container.ReadinessProbe = probe
	// }

	// set how many replicas
	replicas := cr.Spec.CoreServices.OthersBought.Replicas
	// Minimum replicas will be 2
	if replicas == 0 {
		replicas = 2
	}

	// load environment variables and secrets
	optional := true
	// notOptional := false
	container.EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cachev1alpha1.ConfigMapRef,
				},
				Optional: &optional,
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cachev1alpha1.ConfigMapRef,
				},
				Optional: &optional,
			},
		},
	}

	// build and return the othersbought
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Spec.CoreServices.OthersBought.Name,
			Namespace:   cr.Spec.TargetNamespace,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.Spec.CoreServices.OthersBought.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": cr.Spec.CoreServices.OthersBought.Name,
					},
					// Annotations: map[string]string{
					// 	constant.EnvironmentVariablesHashAnnotationKey: cr.Status.EnvironmentVariablesHash,
					// 	constant.SecretVariablesHashAnnotationKey:      cr.Status.SecretVariablesHash,
					// },
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(0),
					Containers: []corev1.Container{
						container,
					},
					Volumes: volumes,
				},
			},
		},
	}
}

// isEnvHashCurrent returns false if the annotation exists and it's value is
// different from the hash provided.
func (d *othersbought) isEnvHashCurrent(current *appsv1.Deployment, annotationKey, hash string) bool {
	if val, ok := current.Spec.Template.Annotations[annotationKey]; ok {
		if val != hash {
			return false
		}
	}
	return true
}

// HasVersionMismatch returns wether the othersbought image is different or not.
func (d *othersbought) HasVersionMismatch(current *appsv1.Deployment, desired *appsv1.Deployment) bool {
	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}
	return false
}

// IsReady returns a true bool if the othersbought has all its pods ready.
func (d *othersbought) IsReady(othersbought *appsv1.Deployment) bool {
	configuredReplicas := othersbought.Status.Replicas
	readyReplicas := othersbought.Status.ReadyReplicas
	othersboughtReady := false
	if configuredReplicas == readyReplicas {
		othersboughtReady = true
	}
	return othersboughtReady
}

//
// Reconcile Functions.
//
func (r *CommerceReconciler) reconcileOthersBought(ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) (ctrl.Result, error) {

	// make sure etcd is ready
	if r.getRunningEtcdPods(cr) != *cr.Spec.Etcd.Replicas {
		err := fmt.Errorf("etcd controllers not ready")
		return ctrl.Result{}, err
	}

	// Define a new Deployment object
	d := NewOthersBought()
	found := &appsv1.Deployment{}
	wanted := d.Create(cr)
	err := r.Get(ctx, types.NamespacedName{Name: wanted.Name, Namespace: wanted.Namespace}, found)

	// Check if this Deployment already exists
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "OthersBought.Namespace", wanted.Namespace, "Deployment.Name", wanted.Name)
		err = r.Create(ctx, wanted)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "OthersBought.Namespace", wanted.Namespace, "Deployment.Name", wanted.Name)
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure othersbought replicas match the desired state
	if !reflect.DeepEqual(found.Spec.Replicas, wanted.Spec.Replicas) {
		log.Info("Current othersbought replicas do not match Deployment configured Replicas")
		// Update the replicas
		err = r.Update(ctx, wanted)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "OthersBought.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
	}

	// Ensure othersbought container image matchs the desired state, returns true if othersbought needs to be updated
	if d.HasVersionMismatch(found, wanted) {
		log.Info("Current othersbought image version do not match Deployment configured version")
		// Update the image
		err = r.Update(ctx, wanted)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "OthersBought.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
	}

	// if !d.isEnvHashCurrent(found, constant.EnvironmentVariablesHashAnnotationKey, cr.Status.EnvironmentVariablesHash) {
	// 	log.Info("Environment variables have been updated")
	// 	patch := client.MergeFrom(found.DeepCopy())
	// 	found.Spec.Template.Annotations[constant.EnvironmentVariablesHashAnnotationKey] = cr.Status.EnvironmentVariablesHash
	// 	if err = r.Patch(ctx, found, patch); err != nil {
	// 		// r.Recorder.Event(found, corev1.EventTypeWarning, "Failed to patch Deployment", fmt.Sprintf("Error: %s", err.Error()))
	// 		r.Log.Error(err, "Failed to patch Deployment")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// if !d.isEnvHashCurrent(found, constant.SecretVariablesHashAnnotationKey, cr.Status.EnvironmentVariablesHash) {
	// 	log.Info("Secret variables have been updated")
	// 	patch := client.MergeFrom(found.DeepCopy())
	// 	found.Spec.Template.Annotations[constant.SecretVariablesHashAnnotationKey] = cr.Status.SecretVariablesHash
	// 	if err = r.Patch(ctx, found, patch); err != nil {
	// 		// r.Recorder.Event(found, corev1.EventTypeWarning, "Failed to patch Deployment", fmt.Sprintf("Error: %s", err.Error()))
	// 		r.Log.Error(err, "Failed to patch Deployment")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// Create list options for listing othersbought pods
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(wanted.Namespace),
		client.MatchingLabels(wanted.Labels),
	}
	err = r.List(ctx, podList, listOpts...)
	if err != nil {
		log.Error(err, "Failed to list Pods.", "OthersBought.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, err
	}

	// Get running Pods from listing above (if any)
	// podNames := getRunningPodNames(podList.Items)

	// Update status.Nodes if needed
	// if !reflect.DeepEqual(podNames, cr.Status.Nodes) || int32(len(podNames)) != cr.Spec.Replicas {
	// 	cr.Status.Nodes = podNames
	// 	// We only want last 8 in status
	// 	if len(cr.Status.Msg) > 7 {
	// 		cr.Status.Msg = append(cr.Status.Msg[len(cr.Status.Msg)-7:], statusMessageFromReplicas(len(podNames), cr.Spec.Replicas))
	// 	} else {
	// 		cr.Status.Msg = append(cr.Status.Msg, statusMessageFromReplicas(len(podNames), cr.Spec.Replicas))
	// 	}

	// 	err := r.Status().Update(ctx, cr)
	// 	if err != nil {
	// 		log.Error(err, "Failed to update status on nodes")
	// 		return ctrl.Result{}, err
	// 	}

	// 	r.Recorder.Event(cr, corev1.EventTypeNormal, "Updated cr.Status", fmt.Sprintf("Status: (%s)", statusMessageFromReplicas(len(podNames), cr.Spec.Replicas)))

	// 	log.Info("Status updated")
	// 	return ctrl.Result{Requeue: true}, nil
	// }

	// Deployment reconcile finished
	return ctrl.Result{}, nil

}