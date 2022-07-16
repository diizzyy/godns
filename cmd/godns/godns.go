package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/TimothyYe/godns/internal/manager"
	"github.com/TimothyYe/godns/internal/settings"
	"github.com/TimothyYe/godns/internal/utils"

	log "github.com/sirupsen/logrus"

	"github.com/fatih/color"
)

var (
	configuration settings.Settings
	optConf       = flag.String("c", "./config.json", "Specify a config file")
	optHelp       = flag.Bool("h", false, "Show help")

	// Version is current version of GoDNS.
	Version = "0.1"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()
	if *optHelp {
		color.Cyan(utils.Logo, Version)
		flag.Usage()
		return
	}

	// Load settings from configurations file
	if err := settings.LoadSettings(*optConf, &configuration); err != nil {
		log.Fatal(err)
	}

	if configuration.DebugInfo {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if err := utils.CheckSettings(&configuration); err != nil {
		log.Fatal("Invalid settings: ", err.Error())
	}

	// start the dns manager
	manager := manager.NewDNSManager(&configuration)
	if err := manager.Run(); err != nil {
		log.Fatal("Failed to start the DNS manager:", err)
	}

	log.Info("GoDNS started...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	manager.Stop()
}
