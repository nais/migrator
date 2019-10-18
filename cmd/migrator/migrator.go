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
	FasitUrl string
}

var (
	cfg = Config{
		FasitUrl: "http://localhost:8080",
	}
	deploy = naisd.Deploy{
		Application: "myapplication",
		Environment: naisd.ENVIRONMENT_P,
		Zone:        naisd.ZONE_FSS,
		Namespace:   "default",
	}
)

func init() {
	flag.StringVar(&deploy.Application, "application", deploy.Application, "application name")
	flag.StringVar(&deploy.Environment, "environment", deploy.Environment, "application environment (p, q, t, u)")
	flag.StringVar(&deploy.Zone, "zone", deploy.Zone, "zone (fss, sbs)")
	flag.StringVar(&deploy.FasitUsername, "fasit-username", deploy.FasitUsername, "fasit username; leave blank to disable")
	flag.StringVar(&deploy.FasitPassword, "fasit-password", deploy.FasitPassword, "fasit password")
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
	var manifest naisd.NaisManifest
	var application naiserator.Application

	decoder := yaml.NewDecoder(os.Stdin)
	err = decoder.Decode(&manifest)
	if err != nil {
		return fmt.Errorf("decode input: %s", err)
	}

	application = mapper.Convert(manifest, deploy)

	encoder := yaml.NewEncoder(os.Stdout)
	err = encoder.Encode(application)
	if err != nil {
		return fmt.Errorf("encode output: %s", err)
	}

	return nil
}
