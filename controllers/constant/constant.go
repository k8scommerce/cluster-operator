package constant

import "time"

const (
	ManagerNamespace      = "k8scom-system"
	TargetNamespace       = "k8scommerce"
	ReconcileRequeueAfter = time.Minute * 10
)
