package commerce

import (
	"context"

	"github.com/go-logr/logr"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	cachev1alpha1 "github.com/localrivet/k8sly-operator/api/v1alpha1"
)

// GatewayIngress interface used to create the kubernetes ingress.
type GatewayIngress interface {
	Create(cr *cachev1alpha1.Commerce) *networking.Ingress
}

// NewIngress instantiates an real implementation of the interface.
func NewGatewayIngress() GatewayIngress {
	return &gatewayIngress{}
}

type gatewayIngress struct{}

// Create returns a new route.
func (r *gatewayIngress) Create(cr *cachev1alpha1.Commerce) *networking.Ingress {
	if cr.Spec.CoreServices.GatewayClient == nil {
		return &networking.Ingress{}
	}

	annotations := map[string]string{
		"kubernetes.io/ingress.class":                          "nginx",
		"nginx.ingress.kubernetes.io/ssl-redirect":             "true",
		"nginx.ingress.kubernetes.io/proxy-body-size":          "5m",
		"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "5m",
		"nginx.org/client-max-body-size":                       "5m",
		"nginx.ingress.kubernetes.io/enable-cors":              "true",
		"nginx.ingress.kubernetes.io/cors-allow-origin":        "*",
		"external-dns.alpha.kubernetes.io/target":              "cr.Spec.Hosts.Client.Hostname",
	}

	hosts := []string{}
	hosts = append(hosts, cr.Spec.Hosts.Client.Hostname)
	// hosts = append(hosts, cr.Spec.Hosts.Admin.Hostname)

	var prefixPathType = networking.PathTypePrefix
	var rules []networking.IngressRule
	for _, host := range hosts {
		rules = append(rules, networking.IngressRule{
			Host: host,
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: []networking.HTTPIngressPath{
						{
							PathType: &prefixPathType,
							Path:     "/",
							Backend: networking.IngressBackend{
								Service: &networking.IngressServiceBackend{
									Name: "gateway-client-service",
									Port: networking.ServiceBackendPort{
										Name: "http",
									},
								},
							},
						},
					},
				},
			},
		})
	}

	ing := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress",
			Namespace: cr.Spec.TargetNamespace,
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "gateway-client-service",
					Port: networking.ServiceBackendPort{
						Name: "http",
					},
				},
			},
			Rules: rules,
		},
	}

	ing.SetAnnotations(annotations)

	// if len(hosts) > 0 {
	// 	ing.Spec.TLS = []networking.IngressTLS{
	// 		{
	// 			Hosts: hosts,
	// 			SecretName: cr.Spec.Urls.Default,
	// 		},
	// 	}
	// }

	return ing

}

func (r *CommerceReconciler) reconcileGatewayIngress(ctx context.Context, cr *cachev1alpha1.Commerce, log logr.Logger) (ctrl.Result, error) {
	i := NewGatewayIngress()
	found := &networking.Ingress{}
	wanted := i.Create(cr)
	err := r.Get(ctx, types.NamespacedName{Name: wanted.Name, Namespace: wanted.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Ingress", "Ingress.Namespace", wanted.Namespace, "Ingress.Name", wanted.Name)
		err = r.Create(ctx, wanted)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Ingress created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Ingress already exists
		log.Info("Ingress already exists", "Ingress.Namespace", found.Namespace, "Ingress.Name", found.Name)
	}

	// Ingress reconcile finished
	return ctrl.Result{}, nil

}
