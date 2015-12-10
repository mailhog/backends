package delivery

import (
	"fmt"
	"os"
	"strings"

	"github.com/mailhog/backends/auth"
	"github.com/mailhog/backends/config"
	"github.com/mailhog/data"
)

// Service represents a delivery service implementation
type Service interface {
	Deliver(msg *data.SMTPMessage) (id string, err error)
	WillDeliver(from, to string, as auth.Identity) bool
}

// Load loads a delivery backend
func Load(cfg config.BackendConfig, appCfg config.AppConfig) Service {
	// FIXME delivery backend could be loaded multiple times, should cache this
	switch strings.ToLower(cfg.Type) {
	case "local":
		return NewLocalDelivery(cfg, appCfg)
	default:
		fmt.Printf("Backend type not recognised\n")
		os.Exit(1)
	}

	return nil
}
