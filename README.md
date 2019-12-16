# migrator: naisd to naiserator migration

Nais daemon is being switched off 02.02.2020. The successor is [Naiserator](https://github.com/nais/naiserator).

Migrator helps with transitioning between systems by converting your `nais.yaml` to Naiserator syntax.

Optionally, it also pulls environment variables and secrets from _Fasit_ to your local drive.

## Usage

Migrator requires Fasit access for most use cases. Running Migrator from your laptop
requires that you set up port forwarding to the Fasit service. This is not generally available
to users and requires port forwarding privileges.

Your best chance is to run Migrator from _utviklerimage_. Use one of our pre-compiled
binaries from the [releases page](https://github.com/nais/migrator/releases/).

```
read fasit_username
read -s fasit_password
./migrator \
    --application myapplication \
    --zone fss \
    --fasit-environment q0 \
    --fasit-username $fasit_username \
    --fasit-password $fasit_password \
    --fasit-url https://fasit.adeo.no \  # only on utviklerimage
    < nais-manifest.yaml \
    > naiserator.yaml
```

If you have port-forwarding capabilities, you can set that up using:

```
kubectl --context prod-fss --namespace default port-forward service/fasit 8080:80
```

## Building

To build a binary, clone the repository and type `make` to compile. You need to [download and install Go](https://golang.org/doc/install) v1.13 or later.

```
git clone https://github.com/nais/migrator
cd migrator
make
# cross compile using `make linux`, `make windows`, `make darwin`
```

## Where to get support

Your first point of information should be the [NAIS user documentation](https://doc.nais.io/observability).

For questions concerning migration in general, [use this Slack thread](https://nav-it.slack.com/archives/C5KUST8N6/p1571300871119200).

To get in touch with the NAIS team, use the [#nais channel](https://nav-it.slack.com/messages/C5KUST8N6) on Slack.

For general discussions between NAIS users and the NAIS team, attend bi-weekly meetings at [NAIS brukerforum](https://nav-it.slack.com/messages/CGGTL83GT).

## Warnings and errors

### Skipping environment variable 'FOO' from secret 'foo'

Please migrate your secrets to Vault.

### Skipping certificate 'foo' in resource 'foo'

This message can probably be ignored, depending on your setup.
Certificate Authority bundles are included automatically in your
application deployment unless using `skipCaBundle: true`.

### Automatic Redis setup is unsupported with Naiserator

If using Redis, it must be deployed as a normal application.
See [Redis on NAIS](https://doc.nais.io/addons/redis) for instructions.

### Alerts must be configured using the Alert resource.

Alerts are no longer part of the NAIS manifest, but has its own resource.
See [custom alerts on NAIS](https://doc.nais.io/observability/alerts) for instructions.
