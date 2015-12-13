package mailbox

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mailhog/backends/config"
	"github.com/mailhog/backends/delivery"
	"github.com/mailhog/backends/resolver"
	"github.com/mailhog/data"
)

// LocalMailbox implements delivery.Service
type LocalMailbox struct {
	maildirPath    string
	maildirPattern string
	cfg            config.BackendConfig
	appCfg         config.AppConfig
	resolver       resolver.Service
}

// UserMailbox represents a users mailbox
type UserMailbox struct {
	mailbox      string
	domain       string
	localMailbox *LocalMailbox
}

// NewLocalMailbox creates a new LocalMailbox backend
func NewLocalMailbox(cfg config.BackendConfig, appCfg config.AppConfig, resolver resolver.Service) Service {
	maildirPath := os.TempDir()
	maildirPattern := "$domain/$mailbox"

	if c, ok := cfg.Data["maildir_path"]; ok {
		if s, ok := c.(string); ok && len(s) > 0 {
			maildirPath = s
		}
	}

	if !strings.HasPrefix(maildirPath, "/") {
		maildirPath = filepath.Join(appCfg.RelPath(), maildirPath)
	}

	err := os.MkdirAll(maildirPath, 0660)
	if err != nil {
		// FIXME
		log.Fatal(err)
	}

	return &LocalMailbox{
		cfg:            cfg,
		appCfg:         appCfg,
		resolver:       resolver,
		maildirPath:    maildirPath,
		maildirPattern: maildirPattern,
	}
}

// Open implements Service.Open
func (l *LocalMailbox) Open(address string) (Mailbox, error) {
	r := l.resolver.Resolve(address)
	log.Printf("%+v", r)
	log.Printf("%+v", resolver.MailboxFound)
	if r.Domain == resolver.DomainPrimaryLocal && r.Mailbox == resolver.MailboxFound {
		p := data.PathFromString(address)
		return &UserMailbox{
			domain:       p.Domain,
			mailbox:      p.Mailbox,
			localMailbox: l,
		}, nil
	}

	return nil, errors.New("mailbox not found")
}

func (m *UserMailbox) filePath() string {
	maildirSuffix := strings.Replace(m.localMailbox.maildirPattern, "$domain", m.domain, -1)
	maildirSuffix = strings.Replace(maildirSuffix, "$mailbox", m.mailbox, -1)
	return filepath.Join(m.localMailbox.maildirPath, maildirSuffix)
}

// Store implements Mailbox.Store
func (m *UserMailbox) Store(msg delivery.Message) error {
	p := m.filePath()
	pt := filepath.Join(p, "tmp")
	pn := filepath.Join(p, "new")

	err := os.MkdirAll(pt, 0660)
	if err != nil {
		return err
	}

	err = os.MkdirAll(pn, 0660)
	if err != nil {
		return err
	}

	opt := filepath.Join(pt, msg.ID+":2,")
	opn := filepath.Join(pn, msg.ID+":2,")

	log.Printf("Writing to %s", opt)

	b := []byte(msg.SMTPMessage.Data)
	err = ioutil.WriteFile(opt, b, 0660)
	if err != nil {
		return err
	}

	return os.Rename(opt, opn)
}
