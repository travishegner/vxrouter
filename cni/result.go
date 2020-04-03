package cni

// Result is a CNI result whos json is printed to STDOUT
type Result struct {
	CniVersion string       `json:"cniVersion"`
	Interfaces []*Interface `json:"interfaces"`
	IPs        []*IP        `json:"ips"`
	Routes     []*Route     `json:"routes"`
	DNS        *DNS         `json:"dns,omitempty"`
}

// Interface represents the created interfaces for the container
type Interface struct {
	Name    string `json:"name"`
	MAC     string `json:"mac,omitempty"`
	Sandbox string `json:"sandbox,omitempty"`
}

// IP is a list of IP configuration information determined by the plugin
type IP struct {
	Version   string `json:"version"`
	Address   string `json:"address"`
	Gateway   string `json:"gateway,omitempty"`
	Interface uint   `json:"interface"`
}

// Route is a route that will be installed into the container's network namespace
type Route struct {
	Destination string `json:"dst"`
	Gateway     string `json:"gw,omitempty"`
}

// DNS represents the information that the container runtime could use to configure the container's DNS information
type DNS struct {
	NameServers   []string `json:"nameservers"`
	Domain        string   `json:"domain"`
	SearchDomains []string `json:"search"`
	Options       []string `json:"options"`
}
