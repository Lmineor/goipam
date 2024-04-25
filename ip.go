package ipam

// IPStorage is a single ipaddress.
type IPStorage struct {
	IP           string `json:"ip"`
	ParentPrefix string `json:"parent_prefix"`
	Namespace    string `json:"namespace"`
}
