package naisd

// naisd deployment request, stripped down
type Deploy struct {
	Application      string `json:"application"`
	Zone             string `json:"zone"`
	FasitEnvironment string `json:"fasitEnvironment,omitempty"`
	FasitUsername    string `json:"fasitUsername,omitempty"`
	FasitPassword    string `json:"fasitPassword,omitempty"`
	Namespace        string `json:"namespace,omitempty"`
}
