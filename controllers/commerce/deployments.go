package commerce

import (
	"fmt"

	cachev1alpha1 "github.com/k8scommerce/cluster-operator/api/v1alpha1"
	"github.com/k8scommerce/cluster-operator/controllers/constant"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewMicroserviceDeployment(ms *cachev1alpha1.MicroService) MicroserviceDeployment {
	return &microServiceDeployment{
		ms: ms,
	}
}

type MicroserviceDeployment interface {
	Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment
}

type microServiceDeployment struct {
	ms *cachev1alpha1.MicroService
}

func (d *microServiceDeployment) Create(cr *cachev1alpha1.Commerce) *appsv1.Deployment {
	microService := d.ms

	if microService == nil {
		return &appsv1.Deployment{}
	}

	annotations := map[string]string{
		"operator-sdk/primary-resource":      fmt.Sprintf("%s/%s", cr.ObjectMeta.Namespace, cr.ObjectMeta.Name),
		"operator-sdk/primary-resource-type": "Commerce.apps",
	}

	// set the default port
	if microService.ContainerPort == 0 {
		microService.ContainerPort = 8080
	}

	// defaults for resource requests
	ensureResourceDefaults(cr, microService)

	if microService.LivenessProbe.Port == 0 {
		microService.LivenessProbe.Port = int(microService.ContainerPort)
	}

	if microService.ReadinessProbe.Port == 0 {
		microService.ReadinessProbe.Port = int(microService.ContainerPort)
	}

	// create the container
	container := corev1.Container{
		Name:  microService.Name,
		Image: CleanContainerImage(microService.Image),
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						"sleep 5",
					},
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: microService.ContainerPort,
				Name:          "http",
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(microService.ReadinessProbe.Port),
				},
			},
			InitialDelaySeconds: microService.ReadinessProbe.InitialDelaySeconds,
			PeriodSeconds:       microService.ReadinessProbe.PeriodSeconds,
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(microService.LivenessProbe.Port),
				},
			},
			InitialDelaySeconds: microService.LivenessProbe.InitialDelaySeconds,
			PeriodSeconds:       microService.LivenessProbe.PeriodSeconds,
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(microService.Resources.Limits.CPU),
				corev1.ResourceMemory: resource.MustParse(microService.Resources.Limits.Memory),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(microService.Resources.Requests.CPU),
				corev1.ResourceMemory: resource.MustParse(microService.Resources.Requests.Memory),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "timezone",
				MountPath: "/etc/localtime",
			},
		},
	}

	volumes := []corev1.Volume{
		{
			Name: "timezone",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/usr/share/zoneinfo/Etc/UTC",
				},
			},
		},
	}

	// does a command exist?
	if len(microService.Command) > 0 {
		container.Command = microService.Command
	}

	if len(microService.Args) > 0 {
		container.Args = microService.Args
	}

	// set how many replicas
	replicas := microService.Replicas
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

	// build and return the cart
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        microService.Name,
			Namespace:   constant.TargetNamespace,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: NewInt32(5),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": microService.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": microService.Name,
					},
					// Annotations: map[string]string{
					// 	constant.EnvironmentVariablesHashAnnotationKey: cr.Status.EnvironmentVariablesHash,
					// 	constant.SecretVariablesHashAnnotationKey:      cr.Status.SecretVariablesHash,
					// },
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(5),
					Containers: []corev1.Container{
						container,
					},
					Volumes: volumes,
				},
			},
		},
	}
}
