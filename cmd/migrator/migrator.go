package main

import (
	"fmt"
	"github.com/nais/migrator/fasit"
	"github.com/nais/migrator/mapper"
	"github.com/nais/migrator/models/naisd"
	"github.com/nais/migrator/models/naiserator"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Config struct {
	FasitURL string
}

var (
	cfg = Config{
		FasitURL: "http://localhost:8080",
	}
	deploy = naisd.Deploy{
		Application:      "myapplication",
		Namespace:        "default",
		Zone:             naisd.ZONE_FSS,
		FasitEnvironment: naisd.ENVIRONMENT_P,
	}
)

func init() {
	flag.StringVar(&deploy.Application, "application", deploy.Application, "application name")
	flag.StringVar(&deploy.Zone, "zone", deploy.Zone, "zone (fss, sbs)")
	flag.StringVar(&cfg.FasitURL, "fasit-url", cfg.FasitURL, "Fasit url")
	flag.StringVar(&deploy.FasitUsername, "fasit-username", deploy.FasitUsername, "Fasit username; leave blank to disable Fasit")
	flag.StringVar(&deploy.FasitPassword, "fasit-password", deploy.FasitPassword, "Fasit password")
	flag.StringVar(&deploy.FasitEnvironment, "fasit-environment", deploy.FasitEnvironment, "Fasit environment ([ptuo][0-9]*")
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
	var fasitResources []fasit.NaisResource

	log.Infoln("Reading NAIS manifest from stdin, hit Ctrl-D when finished")

	decoder := yaml.NewDecoder(os.Stdin)
	err = decoder.Decode(&manifest)
	if err != nil {
		return fmt.Errorf("decode input: %s", err)
	}

	log.Infoln("Finished reading NAIS manifest")

	if len(deploy.FasitUsername) > 0 {
		log.Infof("Fasit integration enabled, retrieving resources for application '%s' environment '%s' zone '%s'\n",
			deploy.Application,
			deploy.FasitEnvironment,
			deploy.Zone,
		)

		fasitClient := fasit.FasitClient{
			FasitUrl: cfg.FasitURL,
			Username: deploy.FasitUsername,
			Password: deploy.FasitPassword,
		}

		timer := time.Now()
		fasitResources, err = fasit.FetchFasitResources(fasitClient, deploy.Application, deploy.FasitEnvironment, deploy.Zone, manifest.FasitResources.Used)
		elapsed := time.Since(timer)

		if err != nil {
			return fmt.Errorf("fetch fasit resources: %s", err)
		}
		log.Infof("Retrieved %d Fasit resources in %s\n", len(fasitResources), elapsed.String())
	}

	application = mapper.Convert(manifest, deploy, fasitResources)

	encoder := yaml.NewEncoder(os.Stdout)
	err = encoder.Encode(application)
	if err != nil {
		return fmt.Errorf("encode output: %s", err)
	}

	return nil
}
