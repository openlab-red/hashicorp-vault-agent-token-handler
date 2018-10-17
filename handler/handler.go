package handler

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron"
)

var log = logrus.New()

func initLog() {
	level, err := logrus.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Level = level
	}

}

func validateConfig() {
	if viper.GetString("vault-capath") == "" {
		log.Fatalln("vault-capath must be set and non-empty")
	}
	if viper.GetString("vault-addr") == "" {
		log.Fatalln("vault-addr must be set and non-empty")
	}
	log.Debugln("configuration is valid")
}

func Start() {
	initLog()
	log.Infoln("Starting token handler...")

	validateConfig()

	tokenHandler := tokenHandler{viper.GetString("vault-addr")}
	newCron(tokenHandler)
	newWatcher(tokenHandler)
}

func newCron(tokenHandler tokenHandler) {
	c := cron.New()
	c.AddFunc(viper.GetString("vault-token-handler-cron"), func() { tokenHandler.readToken() })
	c.Start()
	tokenHandler.readToken()
}

func newWatcher(tokenHandler tokenHandler) {
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorln(err)
	}
	defer watcher.Close()

	//
	done := make(chan bool)

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Infoln(event)
				tokenHandler.readToken()
				// watch for errors
			case err := <-watcher.Errors:
				log.Errorln(err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(viper.GetString("vault-token")); err != nil {
		log.Fatal(err)
	}

	<-done
}
