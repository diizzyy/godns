package manager

import (
	"context"
	"os"

	"github.com/TimothyYe/godns/internal/handler"
	"github.com/TimothyYe/godns/internal/settings"
	"github.com/TimothyYe/godns/internal/utils"
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
	panicChan := make(chan settings.Domain)
	ctx, cancel := context.WithCancel(context.Background())
	d.ctx = ctx
	d.cancelFunc = cancel

	h := handler.CreateHandler(d.settings.Provider)
	h.SetConfiguration(d.settings)
	for _, domain := range d.settings.Domains {
		if d.settings.RunOnce {
			h.DomainLoop(&domain, panicChan, d.settings.RunOnce)
		} else {
			go h.DomainLoop(&domain, panicChan, d.settings.RunOnce)
		}
	}

	if d.settings.RunOnce {
		os.Exit(0)
	}

	panicCount := 0
	for {
		failDomain := <-panicChan
		log.Debug("Got panic in goroutine, will start a new one... :", panicCount)
		go h.DomainLoop(&failDomain, panicChan, d.settings.RunOnce)

		panicCount++
		if panicCount >= utils.PanicMax {
			os.Exit(1)
		}
	}
	return nil
}

func (d *DNSManager) Stop() {
}
