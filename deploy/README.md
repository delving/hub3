# Deploy files

This directory contains the RAPID deployment configuration. This in included in the [goreleaser] [FPM] configuration to build
the RPM and Debian packages.

- systemd/rapid.service = the service definition for systemd. At this time we only support deployments for systems that use [systemd].
- rapid-syslog.conf = the syslog configuration that manages where the RAPID information is logged to.

[goreleaser]: https://github.com/goreleaser/goreleaser
[FPM]: https://github.com/jordansissel/fpm
[systemd]: https://freedesktop.org/wiki/Software/systemd/
