package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/mailhog/backends/config"
	"github.com/mailhog/smtp"
)

/*
  FIXME

  Consider whether mechanisms are defined per-backend or not.

  - Are all (available) mechanisms automatically supported by all backends, e.g. EXTERNAL?
  - Or, are the mechanisms supported specific to a particular backend?
  - Parsing is done in mailhog/smtp, does that make mechanism support a policy decision?
*/

// Service represents an authentication service implementation
type Service interface {
	Authenticate(mechanism string, args ...string) (identity Identity, errorReply *smtp.Reply, ok bool)
	Mechanisms() []string
}

// Identity represents an identity
type Identity interface {
	String() string
	IsValidSender(string) bool
	PolicySet() config.IdentityPolicySet
}

// Load loads an auth backend
func Load(cfg config.BackendConfig, appCfg config.AppConfig) Service {
	// FIXME auth backend could be loaded multiple times, should cache this

	switch strings.ToLower(cfg.Type) {
	case "local":
		return NewLocalAuth(cfg, appCfg)
	default:
		fmt.Printf("Backend type not recognised\n")
		os.Exit(1)
	}

	return nil
}
