package mailbox

import (
	"fmt"
	"os"
	"strings"

	"github.com/mailhog/backends/config"
	"github.com/mailhog/backends/delivery"
	"github.com/mailhog/backends/resolver"
)

// Service represents a delivery service implementation
type Service interface {
	Open(address string) (Mailbox, error)
}

// Mailbox represents a mailbox
type Mailbox interface {
	Store(message delivery.Message) error
}

// Load loads a delivery backend
func Load(cfg config.BackendConfig, appCfg config.AppConfig, resolver resolver.Service) Service {
	// FIXME delivery backend could be loaded multiple times, should cache this
	switch strings.ToLower(cfg.Type) {
	case "local":
		return NewLocalMailbox(cfg, appCfg, resolver)
	default:
		fmt.Printf("Mailbox backend type not recognised\n")
		os.Exit(1)
	}

	return nil
}
