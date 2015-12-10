package delivery

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/fsnotify.v1"

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

func loadSMTPMessage(file string) (*data.SMTPMessage, error) {
	s := &data.SMTPMessage{}

	// FIXME probably inefficient, especially for large files

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	l := strings.Split(string(b), "\n")
	c := 0
	for i, line := range l {
		c = i
		if strings.HasPrefix(line, "HELO:") {
			ln := strings.TrimPrefix(line, "HELO:<")
			ln = strings.TrimSuffix(ln, ">")
			s.Helo = ln
			continue
		}
		if strings.HasPrefix(line, "TO:<") {
			ln := strings.TrimPrefix(line, "TO:<")
			ln = strings.TrimSuffix(ln, ">")
			s.To = append(s.To, ln)
			continue
		}
		if strings.HasPrefix(line, "FROM:<") {
			ln := strings.TrimPrefix(line, "FROM:<")
			ln = strings.TrimSuffix(ln, ">")
			s.From = ln
			continue
		}
		if line == "" {
			break
		}
	}

	s.Data = strings.Join(l[c:], "\n")

	return s, nil
}

// Deliveries implements DeliveryService.Deliveries
func (l *LocalDelivery) Deliveries(c chan *data.SMTPMessage) {
	go func() {
		filepath.Walk(l.spoolNew, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			log.Printf("loading message: %s", path)

			msg, err := loadSMTPMessage(path)
			if err != nil {
				log.Printf("error loading message: %s", err)
				return nil
			}

			c <- msg
			return nil
		})

		var wg sync.WaitGroup

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					log.Println("event:", event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("modified file:", event.Name)
						msg, err := loadSMTPMessage(event.Name)
						if err != nil {
							log.Printf("error loading message: %s", err)
						} else {
							c <- msg
						}
					}
				case err := <-watcher.Errors:
					log.Println("error:", err)
				}
			}
		}()

		err = watcher.Add(l.spoolNew)
		if err != nil {
			log.Fatal(err)
		}
		<-done

		wg.Wait()
	}()
}
