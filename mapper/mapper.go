// package mapper provides a conversion model between naisd and naiserator data types
package mapper

import (
	"fmt"
	"github.com/nais/migrator/models/naisd"
	"github.com/nais/migrator/models/naiserator"
)

func autoIngress(deploy naisd.Deploy) string {
	const format = "https://%s.nais.%s"
	var domain string

	if deploy.Zone == naisd.ZONE_FSS {
		if deploy.Environment == naisd.ENVIRONMENT_P {
			domain = "adeo.no"
		} else {
			domain = "preprod.local"
		}
	} else {
		if deploy.Environment == naisd.ENVIRONMENT_P {
			domain = "oera.no"
		} else {
			domain = "oera-q.local"
		}
	}

	return fmt.Sprintf(format, deploy.Application, domain)
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
func Convert(manifest naisd.NaisManifest, deploy naisd.Deploy) naiserator.Application {
	var ingresses []string

	if !manifest.Ingress.Disabled {
		// TODO: lbconfig from Fasit
		ingresses = append(ingresses, autoIngress(deploy))
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

			// TODO: FasitResources -> environment variables and/or secrets

			LeaderElection: manifest.LeaderElection,

			// TODO: Redis deprecation -> additional application

			// TODO: Alerts deprecation -> alerterator

			Logformat:    manifest.Logformat,
			Logtransform: manifest.Logtransform,
			Vault: naiserator.Vault{
				Enabled: manifest.Secrets,
			},
			WebProxy: manifest.Webproxy,
		},
	}
}
