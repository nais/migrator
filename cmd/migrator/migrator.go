package main

import (
	"fmt"
	"github.com/nais/migrator/mapper"
	"github.com/nais/migrator/models/naisd"
	"github.com/nais/migrator/models/naiserator"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Application   string
	Fasit         bool
	FasitUser     string
	FasitPassword string
	FasitUrl      string
}

var (
	cfg = Config{
		FasitUrl: "http://localhost:8080",
	}
)

func init() {
	flag.StringVar(&cfg.Application, "application", cfg.Application, "application name")
	flag.BoolVar(&cfg.Fasit, "fasit", cfg.Fasit, "use fasit integration")
	flag.StringVar(&cfg.FasitUser, "fasit-user", cfg.FasitUser, "fasit username")
	flag.StringVar(&cfg.FasitPassword, "fasit-password", cfg.FasitPassword, "fasit password")
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: false,
	})
	log.SetOutput(os.Stderr)
	flag.Parse()

	err := run()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func run() error {
	var err error
	var deploy naisd.Deploy
	var manifest naisd.NaisManifest
	var application naiserator.Application

	decoder := yaml.NewDecoder(os.Stdin)
	err = decoder.Decode(&manifest)
	if err != nil {
		return fmt.Errorf("decode input: %s", err)
	}

	deploy.Application = cfg.Application

	application = mapper.Convert(manifest, deploy)

	encoder := yaml.NewEncoder(os.Stdout)
	err = encoder.Encode(application)
	if err != nil {
		return fmt.Errorf("encode output: %s", err)
	}

	return nil
}
