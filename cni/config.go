package cni

// Config represents the cni network config file
type Config struct {
	CniVersion string         `json:"cniVersion"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Args       *Args          `json:"args"`
	Vxlans     []*VxlanConfig `json:"vxlans"`
}

// VxlanConfig represents the configuration for an overlay broadcast domain
type VxlanConfig struct {
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	Cidr         string            `json:"cidr"`
	ExcludeFirst int               `json:"excludeFirst"`
	ExcludeLast  int               `json:"excludeLast"`
	Options      map[string]string `json:"options"`
}

// Args are additional arguments provided by teh container runtime
type Args struct {
	Attributes map[string]interface{} `json:"attributes"`
}
