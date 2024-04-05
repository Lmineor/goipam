package ipam

import (
	"net/netip"
)

// IP is a single ipaddress.
type IP struct {
	IP           netip.Addr `json:"ip"`
	ParentPrefix string     `json:"parent_prefix"`
}

type IPStorage struct {
	IP           string `json:"ip"`
	ParentPrefix string `json:"parent_prefix"`
	Namespace    string `json:"namespace"`
}
