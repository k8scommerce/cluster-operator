package constant

import "time"

const (
	ManagerNamespace      = "k8scommerce-system"
	TargetNamespace       = "k8scommerce"
	ReconcileRequeueAfter = time.Minute * 10
)
