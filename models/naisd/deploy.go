package naisd

// naisd deployment request
type Deploy struct {
	Application      string `json:"application"`
	Version          string `json:"version"`
	Zone             string `json:"zone"`
	ManifestUrl      string `json:"manifesturl,omitempty"`
	SkipFasit        bool   `json:"skipFasit,omitempty"`
	FasitEnvironment string `json:"fasitEnvironment,omitempty"`
	FasitUsername    string `json:"fasitUsername,omitempty"`
	FasitPassword    string `json:"fasitPassword,omitempty"`
	OnBehalfOf       string `json:"onbehalfof,omitempty"`
	Namespace        string `json:"namespace,omitempty"`
	Environment      string `json:"environment,omitempty"`
}
