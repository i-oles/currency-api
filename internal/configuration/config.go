package configuration

import (
	"encoding/json"
	"flag"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/tkanos/gonfig"
)

const (
	defaultCfgFilename = "dev.json"
)

type Configuration struct {
	ListenAddress  string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ContextTimeout time.Duration
	APIURL         string
}

func (c *Configuration) Pretty() string {
	cfgPretty, _ := json.MarshalIndent(c, "", "  ")

	return string(cfgPretty)
}

func GetConfig(cfgPath string, cfg *Configuration) error {
	cfgFilename := flag.String("cfgFile", "", "config filename to load")
	flag.Parse()

	config := defaultCfgFilename
	if *cfgFilename != "" {
		config = *cfgFilename
	}

	cfgFinalPath := filepath.Join(cfgPath, config)

	err := gonfig.GetConf(cfgFinalPath, cfg)
	if err != nil {
		slog.Error("config error", slog.String("err", err.Error()))
	}

	return nil
}
