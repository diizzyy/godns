package manager

import (
	"context"
	"os"

	"github.com/TimothyYe/godns/internal/handler"
	"github.com/TimothyYe/godns/internal/settings"
	log "github.com/sirupsen/logrus"
)

type DNSManager struct {
	settings   *settings.Settings
	cancelFunc context.CancelFunc
	ctx        context.Context
}

func NewDNSManager(settings *settings.Settings) *DNSManager {
	return &DNSManager{
		settings: settings,
	}
}

func (d *DNSManager) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	d.ctx = ctx
	d.cancelFunc = cancel

	h := handler.CreateHandler(d.settings.Provider)
	h.SetConfiguration(d.settings)
	for _, domain := range d.settings.Domains {
		if d.settings.RunOnce {
			h.DomainLoop(&domain, d.settings.RunOnce)
		} else {
			go h.DomainLoop(&domain, d.settings.RunOnce)
		}
	}

	if d.settings.RunOnce {
		os.Exit(0)
	}

	return nil
}

func (d *DNSManager) Stop() {
	log.Info("Terminating GoDNS...")
}
