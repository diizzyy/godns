package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/TimothyYe/godns/internal/provider"

	log "github.com/sirupsen/logrus"

	"github.com/TimothyYe/godns/internal/settings"
	"github.com/TimothyYe/godns/internal/utils"
	"github.com/TimothyYe/godns/pkg/lib"
	"github.com/TimothyYe/godns/pkg/notification"
)

type Handler struct {
	conf                *settings.Settings
	ctx                 context.Context
	dnsProvider         provider.IDNSProvider
	notificationManager notification.INotificationManager
	cachedIP            string
}

func NewHandler(ctx context.Context, conf *settings.Settings) Handler {
	handler := Handler{
		conf: conf,
		ctx:  ctx,
	}

	handler.notificationManager = notification.GetNotificationManager(handler.conf)
	return handler
}

func (handler *Handler) SetProvider(provider provider.IDNSProvider) {
	handler.dnsProvider = provider
}

func (handler *Handler) DomainLoop(domain *settings.Domain, runOnce bool) {
	for while := true; while; while = !runOnce {
		handler.domainLoop(domain)

		if runOnce {
			break
		}

		log.Debugf("DNS update loop finished, will run again in %d seconds", handler.conf.Interval)
		time.Sleep(time.Second * time.Duration(handler.conf.Interval))
	}
}

func (handler *Handler) domainLoop(domain *settings.Domain) {
	ip, err := utils.GetCurrentIP(handler.conf)
	if err != nil {
		log.Error(err)
		return
	}
	if ip == handler.cachedIP {
		log.Debugf("IP (%s) matches cached IP (%s), skipping", ip, handler.cachedIP)
		return
	}
	err = handler.updateDNS(domain, ip)
	if err != nil {
		log.Error(err)
		return
	}
	handler.cachedIP = ip
	log.Debugf("Cached IP address: %s", ip)
}

func (handler *Handler) updateDNS(domain *settings.Domain, ip string) error {
	for _, subdomainName := range domain.SubDomains {

		var hostname string
		if subdomainName != utils.RootDomain {
			hostname = subdomainName + "." + domain.DomainName
		} else {
			hostname = domain.DomainName
		}

		lastIP, err := utils.ResolveDNS(hostname, handler.conf.Resolver, handler.conf.IPType)
		if err != nil {
			log.Error(err)
			continue
		}

		//check against the current known IP, if no change, skip update
		if ip == lastIP {
			log.Infof("IP is the same as cached one (%s). Skip update.", ip)
		} else {
			if err := handler.dnsProvider.UpdateIP(domain.DomainName, subdomainName, ip); err != nil {
				return err
			}

			successMessage := fmt.Sprintf("%s.%s", subdomainName, domain.DomainName)
			handler.notificationManager.Send(successMessage, ip)

			// execute webhook when it is enabled
			if handler.conf.Webhook.Enabled {
				if err := lib.GetWebhook(handler.conf).Execute(hostname, ip); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
