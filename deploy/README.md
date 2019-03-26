# Deploy files

This directory contains the Hub3 deployment configuration. This in included in the [goreleaser] [FPM] configuration to build
the RPM and Debian packages.

- systemd/hub3.service = the service definition for systemd. At this time we only support deployments for systems that use [systemd].
- hub3-syslog.conf = the syslog configuration that manages where the Hub3 information is logged to.

[goreleaser]: https://github.com/goreleaser/goreleaser
[FPM]: https://github.com/jordansissel/fpm
[systemd]: https://freedesktop.org/wiki/Software/systemd/
