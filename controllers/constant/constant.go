package constant

import "time"

const (
	// ReconcileRequeueAfter on success we will requeue on this interval so
	// secrets can be polled and updated if they have changed.
	ReconcileRequeueAfter = time.Minute * 10
	// A single cerberus call will be allowed within this time interval.
	CerberusCallTimeInterval = time.Minute * 10
	// UseExistingCluster allows the tests to run in KinD.
	UseExistingCluster = "TEST_WITH_EXISTING_CLUSTER"
	// EventStatusNamespace defines the namespace for saving events.
	EventStatusNamespace = "default"
	// SecretsOkta is the name of the secrets file that contains the okta client credentials.
	SecretsOkta = "client-credentials"
	// SecretsOktaNamespace is the namespace where the okta secrets file is located.
	SecretsOktaNamespace = "npe-system"
	// StateStart defines the start state value.
	StateStart = "START"
	// StateFinal defines the final state value.
	StateFinal = "FINAL"
	// NamespacePrefix is the prefix used when creating the namespace for the environment (e.g. tenant-{uuid}).
	NamespacePrefix = "tenant-"
	// SecretsName is the name of the secrets file used to stored sensitive environment variables.
	SecretsName = "secret-environment-variables"
	// EnvironmentVariables is the name of the config map used to store environment variables.
	EnvironmentVariables = "environment-variables"
	// EnvironmentVariablesHashAnnotationKey is the annotation key storing the hash of the env vars causing container restart on change.
	EnvironmentVariablesHashAnnotationKey = "environmentVariablesHash"
	// SecretVariablesHashAnnotationKey is the annotation key storing the hash of the env vars from secrets causing container restart on change.
	SecretVariablesHashAnnotationKey = "secretVariablesHash"
	// ServiceAccountUserName is the name of the environment variable that stores the service account username for tests.
	ServiceAccountUserName = "SERVICE_ACCOUNT_USERNAME"
	// ServiceAccountPassword is the name of the environment variable that stores the service account password for tests.
	ServiceAccountPassword = "SERVICE_ACCOUNT_PASSWORD"
)
