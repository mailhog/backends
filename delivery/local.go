package delivery

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mailhog/backends/auth"
	"github.com/mailhog/backends/config"
	"github.com/mailhog/data"
)

// LocalDelivery implements delivery.Service
type LocalDelivery struct {
	spoolPath string
	spoolTmp  string
	spoolNew  string
	cfg       config.BackendConfig
	appCfg    config.AppConfig
}

// NewLocalDelivery creates a new LocalDelivery backend
func NewLocalDelivery(cfg config.BackendConfig, appCfg config.AppConfig) Service {
	spoolPath := os.TempDir()

	if c, ok := cfg.Data["spool_path"]; ok {
		if s, ok := c.(string); ok && len(s) > 0 {
			spoolPath = s
		}
	}

	if !strings.HasPrefix(spoolPath, "/") {
		spoolPath = filepath.Join(appCfg.RelPath(), spoolPath)
	}

	spoolTmp := filepath.Join(spoolPath, "tmp")
	spoolNew := filepath.Join(spoolPath, "new")

	err := os.MkdirAll(spoolPath, 0660)
	if err != nil {
		// FIXME
		log.Fatal(err)
	}

	err = os.MkdirAll(spoolTmp, 0660)
	if err != nil {
		// FIXME
		log.Fatal(err)
	}

	err = os.MkdirAll(spoolNew, 0660)
	if err != nil {
		// FIXME
		log.Fatal(err)
	}

	return &LocalDelivery{
		cfg:       cfg,
		appCfg:    appCfg,
		spoolPath: spoolPath,
		spoolTmp:  spoolTmp,
		spoolNew:  spoolNew,
	}
}

// Deliver implements DeliveryService.Deliver
func (l *LocalDelivery) Deliver(msg *data.SMTPMessage) (id string, err error) {
	var mid data.MessageID

	// FIXME this is for storage, so isn't strictly the "Message-ID"
	// as defined by the message header, or what the data.NewMessageID function
	// was intended for.
	mid, err = data.NewMessageID("FIXME")
	if err != nil {
		return
	}
	id = string(mid)

	dpTmp := filepath.Join(l.spoolTmp, id)
	dpNew := filepath.Join(l.spoolNew, id)

	b, err := ioutil.ReadAll(msg.Bytes())
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(dpTmp, b, 0660)

	if err == nil {
		err = os.Rename(dpTmp, dpNew)
	}

	return
}

// WillDeliver implements DeliveryService.WillDeliver
func (l *LocalDelivery) WillDeliver(from, to string, as auth.Identity) bool {
	return true
}
