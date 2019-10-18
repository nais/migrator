// Package naisd contains the data model for the nais daemon manifest - "legacy format"
package naisd

type Probe struct {
	Path             string
	InitialDelay     int `yaml:"initialDelay"`
	PeriodSeconds    int `yaml:"periodSeconds"`
	FailureThreshold int `yaml:"failureThreshold"`
	Timeout          int `yaml:"timeout"`
}

type Healthcheck struct {
	Liveness  Probe
	Readiness Probe
}

type ResourceList struct {
	Cpu    string
	Memory string
}

type ResourceRequirements struct {
	Limits   ResourceList
	Requests ResourceList
}

type PrometheusConfig struct {
	Enabled bool
	Port    string
	Path    string
}

type IstioConfig struct {
	Enabled bool
}

type Vault struct {
	Enabled bool
	Sidecar bool
}

type NaisManifest struct {
	Team               string
	Image              string
	Port               int
	DeploymentStrategy string
	Healthcheck        Healthcheck
	PreStopHookPath    string `yaml:"preStopHookPath"`
	Prometheus         PrometheusConfig
	Istio              IstioConfig
	Replicas           Replicas
	Ingress            Ingress
	Resources          ResourceRequirements
	FasitResources     FasitResources `yaml:"fasitResources"`
	LeaderElection     bool           `yaml:"leaderElection"`
	Redis              Redis          `yaml:"redis"`
	Alerts             []PrometheusAlertRule
	Logformat          string
	Logtransform       string
	Secrets            bool `yaml:"secrets"`
	Vault              Vault
	Webproxy           bool `yaml:"webproxy"`
}

type Ingress struct {
	Disabled bool
}

type Replicas struct {
	Min                    int
	Max                    int
	CpuThresholdPercentage int `yaml:"cpuThresholdPercentage"`
}

type FasitResources struct {
	Used    []UsedResource
	Exposed []ExposedResource
}

type UsedResource struct {
	Alias        string
	ResourceType string            `yaml:"resourceType"`
	PropertyMap  map[string]string `yaml:"propertyMap"`
}

type ExposedResource struct {
	Alias          string
	ResourceType   string `yaml:"resourceType"`
	Path           string
	Description    string
	WsdlGroupId    string `yaml:"wsdlGroupId"`
	WsdlArtifactId string `yaml:"wsdlArtifactId"`
	WsdlVersion    string `yaml:"wsdlVersion"`
	SecurityToken  string `yaml:"securityToken"`
	AllZones       bool   `yaml:"allZones"`
}

type Redis struct {
	Enabled  bool
	Image    string
	Limits   ResourceList
	Requests ResourceList
}

type PrometheusAlertRule struct {
	Alert       string
	Expr        string
	For         string
	Labels      map[string]string
	Annotations map[string]string
}
