// package mapper provides a conversion model between naisd and naiserator data types
package mapper

import (
	"fmt"
	"github.com/nais/migrator/fasit"
	"github.com/nais/migrator/models/naisd"
	"github.com/nais/migrator/models/naiserator"
	log "github.com/sirupsen/logrus"
	"net/url"
)

func autoIngress(deploy naisd.Deploy) string {
	const format = "https://%s.nais.%s"
	var domain string

	if deploy.Zone == naisd.ZONE_FSS {
		if deploy.FasitEnvironment == naisd.ENVIRONMENT_P {
			domain = "adeo.no"
		} else {
			domain = "preprod.local"
		}
	} else {
		if deploy.FasitEnvironment == naisd.ENVIRONMENT_P {
			domain = "oera.no"
		} else {
			domain = "oera-q.local"
		}
	}

	return fmt.Sprintf(format, deploy.Application, domain)
}

func fasitIngress(resources []fasit.NaisResource) []string {
	var ingresses []string
	var u url.URL

	for _, resource := range resources {
		for _, ingress := range resource.Ingresses {
			u.Scheme = "https"
			u.Path = ingress.Path
			u.Host = ingress.Host

			ingresses = append(ingresses, u.String())
		}
	}

	return ingresses
}

func fasitEnv(resources []fasit.NaisResource) []naiserator.EnvVar {
	var vars []naiserator.EnvVar

	// TODO: REDIS_HOST with redis:true

	for _, resource := range resources {
		for k := range resource.Secret {
			log.Warnf("Skipping environment variable '%s' from secret '%s'", resource.ToEnvironmentVariable(k), resource.Name)
		}
		for k := range resource.Certificates {
			log.Warnf("Skipping certificate '%s' in resource '%s'", k, resource.Name)
			if resource.Name == "nav_truststore" {
				log.Infof("Certificate in resource '%s' is automatically included in Naiserator deployments", k)
			}
		}
		for key, val := range resource.Properties {
			vars = append(vars, naiserator.EnvVar{
				Name:  resource.ToEnvironmentVariable(key),
				Value: val,
			})
		}
	}

	return vars
}

func probeConvert(manifest naisd.NaisManifest, probe naisd.Probe) naiserator.Probe {
	return naiserator.Probe{
		Port:             manifest.Port,
		Path:             probe.Path,
		Timeout:          probe.Timeout,
		FailureThreshold: probe.FailureThreshold,
		PeriodSeconds:    probe.PeriodSeconds,
		InitialDelay:     probe.InitialDelay,
	}
}

func prometheusConvert(config naisd.PrometheusConfig) naiserator.PrometheusConfig {
	return naiserator.PrometheusConfig{
		Path:    config.Path,
		Port:    config.Port,
		Enabled: config.Enabled,
	}
}

func replicaConvert(replicas naisd.Replicas) naiserator.Replicas {
	return naiserator.Replicas{
		Min:                    replicas.Min,
		Max:                    replicas.Max,
		CpuThresholdPercentage: replicas.CpuThresholdPercentage,
	}
}

func resourceConvert(config naisd.ResourceList) naiserator.ResourceSpec {
	return naiserator.ResourceSpec{
		Cpu:    config.Cpu,
		Memory: config.Memory,
	}
}

// Convert from naisd manifest to Naiserator application Kubernetes resource.
func Convert(manifest naisd.NaisManifest, deploy naisd.Deploy, resources []fasit.NaisResource) naiserator.Application {
	var ingresses []string

	if !manifest.Ingress.Disabled {
		ingresses = append(ingresses, autoIngress(deploy))
		ingresses = append(ingresses, fasitIngress(resources)...)
	}

	// TODO: fix automatically by creating another Application spec?
	if manifest.Redis.Enabled {
		log.Warn("Automatic Redis setup is unsupported with Naiserator.")
	}

	// TODO: fix automatically by creating an Alert spec?
	if len(manifest.Alerts) > 0 {
		log.Warn("Alerts must be configured using the Alert resource.")
	}

	return naiserator.Application{
		TypeMeta: naiserator.TypeMeta{
			Kind:       "Application",
			APIVersion: "nais.io/v1alpha1",
		},
		ObjectMeta: naiserator.ObjectMeta{
			Name: deploy.Application,
			Labels: map[string]string{
				"team": manifest.Team,
			},
			Namespace: deploy.Namespace,
		},
		Spec: naiserator.ApplicationSpec{
			Image: manifest.Image,
			Port:  manifest.Port,
			Strategy: naiserator.Strategy{
				Type: manifest.DeploymentStrategy,
			},
			Readiness:       probeConvert(manifest, manifest.Healthcheck.Readiness),
			Liveness:        probeConvert(manifest, manifest.Healthcheck.Liveness),
			PreStopHookPath: manifest.PreStopHookPath,
			Prometheus:      prometheusConvert(manifest.Prometheus),
			Replicas:        replicaConvert(manifest.Replicas),
			Ingresses:       ingresses,
			Resources: naiserator.ResourceRequirements{
				Requests: resourceConvert(manifest.Resources.Requests),
				Limits:   resourceConvert(manifest.Resources.Limits),
			},
			Env:            fasitEnv(resources),
			LeaderElection: manifest.LeaderElection,

			Logformat:    manifest.Logformat,
			Logtransform: manifest.Logtransform,
			Vault: naiserator.Vault{
				Enabled: manifest.Secrets,
			},
			WebProxy: manifest.Webproxy,
		},
	}
}
