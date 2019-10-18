# migrator: naisd to naiserator migration

Nais daemon is being switched off 02.02.2020. The successor is [Naiserator](https://github.com/nais/naiserator).

Migrator helps with transitioning between systems by converting your `nais.yaml` to Naiserator syntax.

Optionally, it also pulls environment variables and secrets from _Fasit_ to your local drive.

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
