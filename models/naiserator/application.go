package naiserator

// +groupName="nais.io"

// Application defines a NAIS application.
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Team",type="string",JSONPath=".metadata.labels.team"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.deploymentRolloutStatus"
// +kubebuilder:resource:path="applications",shortName="app",singular="application"
type Application struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata,omitempty"`

	Spec   ApplicationSpec   `yaml:"spec"`
	Status ApplicationStatus `yaml:"status,omitempty"`
}

// ApplicationSpec contains the NAIS manifest.
type ApplicationSpec struct {
	AccessPolicy    AccessPolicy         `yaml:"accessPolicy,omitempty"`
	ConfigMaps      ConfigMaps           `yaml:"configMaps,omitempty"`
	Env             []EnvVar             `yaml:"env,omitempty"`
	Image           string               `yaml:"image"`
	Ingresses       []string             `yaml:"ingresses,omitempty"`
	LeaderElection  bool                 `yaml:"leaderElection,omitempty"`
	Liveness        Probe                `yaml:"liveness,omitempty"`
	Logtransform    string               `yaml:"logtransform,omitempty"`
	Port            int                  `yaml:"port,omitempty"`
	PreStopHookPath string               `yaml:"preStopHookPath,omitempty"`
	Prometheus      PrometheusConfig     `yaml:"prometheus,omitempty"`
	Readiness       Probe                `yaml:"readiness,omitempty"`
	Replicas        Replicas             `yaml:"replicas,omitempty"`
	Resources       ResourceRequirements `yaml:"resources,omitempty"`
	Secrets         []Secret             `yaml:"secrets,omitempty"`
	SecureLogs      SecureLogs           `yaml:"secureLogs,omitempty"`
	Service         Service              `yaml:"service,omitempty"`
	SkipCaBundle    bool                 `yaml:"skipCaBundle,omitempty"`
	Strategy        Strategy             `yaml:"strategy,omitempty"`
	Vault           Vault                `yaml:"vault,omitempty"`
	WebProxy        bool                 `yaml:"webproxy,omitempty"`

	// +kubebuilder:validation:Enum="";accesslog;accesslog_with_processing_time;accesslog_with_referer_useragent;capnslog;logrus;gokit;redis;glog;simple;influxdb;log15
	Logformat string `yaml:"logformat,omitempty"`
}

// ApplicationStatus contains different NAIS status properties
type ApplicationStatus struct {
	CorrelationID           string `yaml:"correlationID,omitempty"`
	DeploymentRolloutStatus string `yaml:"deploymentRolloutStatus,omitempty"`
}

type SecureLogs struct {
	// Whether or not to enable a sidecar container for secure logging.
	Enabled bool `yaml:"enabled"`
}

// Liveness probe and readiness probe definitions.
type Probe struct {
	Path             string `yaml:"path"`
	Port             int    `yaml:"port,omitempty"`
	InitialDelay     int    `yaml:"initialDelay,omitempty"`
	PeriodSeconds    int    `yaml:"periodSeconds,omitempty"`
	FailureThreshold int    `yaml:"failureThreshold,omitempty"`
	Timeout          int    `yaml:"timeout,omitempty"`
}

type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled,omitempty"`
	Port    string `yaml:"port,omitempty"`
	Path    string `yaml:"path,omitempty"`
}

type Replicas struct {
	// The minimum amount of replicas acceptable for a successful deployment.
	Min int `yaml:"min,omitempty"`
	// The pod autoscaler will scale deployments on demand until this maximum has been reached.
	Max int `yaml:"max,omitempty"`
	// Amount of CPU usage before the autoscaler kicks in.
	CpuThresholdPercentage int `yaml:"cpuThresholdPercentage,omitempty"`
}

type ResourceSpec struct {
	// +kubebuilder:validation:Pattern=^\d+m?$
	Cpu string `yaml:"cpu,omitempty"`
	// +kubebuilder:validation:Pattern=^\d+[KMG]i$
	Memory string `yaml:"memory,omitempty"`
}

type ResourceRequirements struct {
	Limits   ResourceSpec `yaml:"limits,omitempty"`
	Requests ResourceSpec `yaml:"requests,omitempty"`
}

type ObjectFieldSelector struct {
	// +kubebuilder:validation:Enum="";metadata.name;metadata.namespace;metadata.labels;metadata.annotations;spec.nodeName;spec.serviceAccountName;status.hostIP;status.podIP
	FieldPath string `yaml:"fieldPath"`
}

type EnvVarSource struct {
	FieldRef ObjectFieldSelector `yaml:"fieldRef"`
}

type EnvVar struct {
	Name      string       `yaml:"name"`
	Value     string       `yaml:"value,omitempty"`
	ValueFrom EnvVarSource `yaml:"valueFrom,omitempty"`
}

type SecretPath struct {
	MountPath string `yaml:"mountPath"`
	KvPath    string `yaml:"kvPath"`
}

type Vault struct {
	Enabled bool         `yaml:"enabled,omitempty"`
	Sidecar bool         `yaml:"sidecar,omitempty"`
	Mounts  []SecretPath `yaml:"paths,omitempty"`
}

type ConfigMaps struct {
	Files []string `yaml:"files,omitempty"`
}

type Strategy struct {
	// +kubebuilder:validation:Enum=Recreate;RollingUpdate
	Type string `yaml:"type"`
}

type Service struct {
	Port int32 `yaml:"port"`
}

type AccessPolicyExternalRule struct {
	Host string `yaml:"host"`
}

type AccessPolicyGressRule struct {
	Application string `yaml:"application"`
	Namespace   string `yaml:"namespace,omitempty"`
}

type AccessPolicyInbound struct {
	Rules []AccessPolicyGressRule `yaml:"rules"`
}

type AccessPolicyOutbound struct {
	Rules    []AccessPolicyGressRule    `yaml:"rules"`
	External []AccessPolicyExternalRule `yaml:"external"`
}

type AccessPolicy struct {
	Inbound  AccessPolicyInbound  `yaml:"inbound,omitempty"`
	Outbound AccessPolicyOutbound `yaml:"outbound,omitempty"`
}

type Secret struct {
	Name string `yaml:"name"`
	// +kubebuilder:validation:Enum="";env;files
	Type      string `yaml:"type,omitempty"`
	MountPath string `yaml:"mountPath,omitempty"`
}

type ObjectMeta struct {
	Name        string            `yaml:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Namespace   string            `yaml:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	Labels      map[string]string `yaml:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
	Annotations map[string]string `yaml:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

type TypeMeta struct {
	Kind       string `yaml:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	APIVersion string `yaml:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}
