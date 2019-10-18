package naiserator

// +groupName="nais.io"

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Application defines a NAIS application.
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Team",type="string",JSONPath=".metadata.labels.team"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.deploymentRolloutStatus"
// +kubebuilder:resource:path="applications",shortName="app",singular="application"
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// ApplicationSpec contains the NAIS manifest.
type ApplicationSpec struct {
	AccessPolicy    AccessPolicy         `json:"accessPolicy,omitempty"`
	ConfigMaps      ConfigMaps           `json:"configMaps,omitempty"`
	Env             []EnvVar             `json:"env,omitempty"`
	Image           string               `json:"image"`
	Ingresses       []string             `json:"ingresses,omitempty"`
	LeaderElection  bool                 `json:"leaderElection,omitempty"`
	Liveness        Probe                `json:"liveness,omitempty"`
	Logtransform    string               `json:"logtransform,omitempty"`
	Port            int                  `json:"port,omitempty"`
	PreStopHookPath string               `json:"preStopHookPath,omitempty"`
	Prometheus      PrometheusConfig     `json:"prometheus,omitempty"`
	Readiness       Probe                `json:"readiness,omitempty"`
	Replicas        Replicas             `json:"replicas,omitempty"`
	Resources       ResourceRequirements `json:"resources,omitempty"`
	Secrets         []Secret             `json:"secrets,omitempty"`
	SecureLogs      SecureLogs           `json:"secureLogs,omitempty"`
	Service         Service              `json:"service,omitempty"`
	SkipCaBundle    bool                 `json:"skipCaBundle,omitempty"`
	Strategy        Strategy             `json:"strategy,omitempty"`
	Vault           Vault                `json:"vault,omitempty"`
	WebProxy        bool                 `json:"webproxy,omitempty"`

	// +kubebuilder:validation:Enum="";accesslog;accesslog_with_processing_time;accesslog_with_referer_useragent;capnslog;logrus;gokit;redis;glog;simple;influxdb;log15
	Logformat string `json:"logformat,omitempty"`
}

// ApplicationStatus contains different NAIS status properties
type ApplicationStatus struct {
	CorrelationID           string `json:"correlationID,omitempty"`
	DeploymentRolloutStatus string `json:"deploymentRolloutStatus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Application `json:"items"`
}

type SecureLogs struct {
	// Whether or not to enable a sidecar container for secure logging.
	Enabled bool `json:"enabled"`
}

// Liveness probe and readiness probe definitions.
type Probe struct {
	Path             string `json:"path"`
	Port             int    `json:"port,omitempty"`
	InitialDelay     int    `json:"initialDelay,omitempty"`
	PeriodSeconds    int    `json:"periodSeconds,omitempty"`
	FailureThreshold int    `json:"failureThreshold,omitempty"`
	Timeout          int    `json:"timeout,omitempty"`
}

type PrometheusConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Port    string `json:"port,omitempty"`
	Path    string `json:"path,omitempty"`
}

type Replicas struct {
	// The minimum amount of replicas acceptable for a successful deployment.
	Min int `json:"min,omitempty"`
	// The pod autoscaler will scale deployments on demand until this maximum has been reached.
	Max int `json:"max,omitempty"`
	// Amount of CPU usage before the autoscaler kicks in.
	CpuThresholdPercentage int `json:"cpuThresholdPercentage,omitempty"`
}

type ResourceSpec struct {
	// +kubebuilder:validation:Pattern=^\d+m?$
	Cpu string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Pattern=^\d+[KMG]i$
	Memory string `json:"memory,omitempty"`
}

type ResourceRequirements struct {
	Limits   ResourceSpec `json:"limits,omitempty"`
	Requests ResourceSpec `json:"requests,omitempty"`
}

type ObjectFieldSelector struct {
	// +kubebuilder:validation:Enum="";metadata.name;metadata.namespace;metadata.labels;metadata.annotations;spec.nodeName;spec.serviceAccountName;status.hostIP;status.podIP
	FieldPath string `json:"fieldPath"`
}

type EnvVarSource struct {
	FieldRef ObjectFieldSelector `json:"fieldRef"`
}

type EnvVar struct {
	Name      string       `json:"name"`
	Value     string       `json:"value,omitempty"`
	ValueFrom EnvVarSource `json:"valueFrom,omitempty"`
}

type SecretPath struct {
	MountPath string `json:"mountPath"`
	KvPath    string `json:"kvPath"`
}

type Vault struct {
	Enabled bool         `json:"enabled,omitempty"`
	Sidecar bool         `json:"sidecar,omitempty"`
	Mounts  []SecretPath `json:"paths,omitempty"`
}

type ConfigMaps struct {
	Files []string `json:"files,omitempty"`
}

type Strategy struct {
	// +kubebuilder:validation:Enum=Recreate;RollingUpdate
	Type string `json:"type"`
}

type Service struct {
	Port int32 `json:"port"`
}

type AccessPolicyExternalRule struct {
	Host string `json:"host"`
}

type AccessPolicyGressRule struct {
	Application string `json:"application"`
	Namespace   string `json:"namespace,omitempty"`
}

type AccessPolicyInbound struct {
	Rules []AccessPolicyGressRule `json:"rules"`
}

type AccessPolicyOutbound struct {
	Rules    []AccessPolicyGressRule    `json:"rules"`
	External []AccessPolicyExternalRule `json:"external"`
}

type AccessPolicy struct {
	Inbound  AccessPolicyInbound  `json:"inbound,omitempty"`
	Outbound AccessPolicyOutbound `json:"outbound,omitempty"`
}

type Secret struct {
	Name string `json:"name"`
	// +kubebuilder:validation:Enum="";env;files
	Type      string `json:"type,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
}
