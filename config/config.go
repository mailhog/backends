package config

// BackendConfig defines an individual backend configuration
type BackendConfig struct {
	Type string                 `json:",omitempty"`
	Ref  string                 `json:",omitempty"`
	Data map[string]interface{} `json:",omitempty"`
}

// AppConfig defines the application configuration required by backends
type AppConfig interface {
	RelPath() string
}

// IdentityPolicySet defines the policies which can be applied per-user
type IdentityPolicySet struct {
	RequireLocalDelivery    *bool
	MaximumRecipients       *int
	RejectInvalidRecipients *bool
}

// DefaultIdentityPolicySet defines a default policy set with no non-nil options
func DefaultIdentityPolicySet() IdentityPolicySet {
	return IdentityPolicySet{}
}
