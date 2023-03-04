package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

type config struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	SSL      bool   `json:"-"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	BaseUrl  string `json:"baseUrl"`
	HomePath string `json:"homePath"`
}

func readConfig() *config {
	w, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Unable to read config.json: %v", err)
	}
	cfg := &config{}
	err = json.Unmarshal(w, cfg)
	if err != nil {
		log.Fatalf("Unable to parse config.json: %v", err)
	}
	if len(cfg.Hostname) == 0 {
		log.Fatalf("config: hostname not set")
	}
	if cfg.Port == 0 {
		log.Fatalf("config: keyFile not set")
	}
	if len(cfg.CertFile) > 0 || len(cfg.KeyFile) > 0 {
		cfg.SSL = true
	}
	if cfg.SSL {
		if len(cfg.CertFile) == 0 {
			log.Fatalf("config: certFile not set")
		}
		if len(cfg.KeyFile) == 0 {
			log.Fatalf("config: keyFile not set")
		}
	}
	if len(cfg.BaseUrl) == 0 {
		log.Fatalf("config: baseUrl not set")
	}
	if len(cfg.HomePath) == 0 {
		log.Fatalf("config: homePath not set")
	}
	return cfg
}
