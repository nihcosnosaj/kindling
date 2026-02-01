## Kindling

Kindling is a CLI tool I use for local testing and development with Kubernetes since I'm on a mac at home. My linux homelab machine(s) is/are still in development. :-) All it does is, based on CLI args, boots up an ephemeral k8s cluster locally using `kind` (Kubernetes in Docker)

[![CI](https://github.com/YOUR_USER/kindling/actions/workflows/ci.yml/badge.svg)](https://github.com/nihcosnosaj/kindling/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USER/kindling)](https://goreportcard.com/report/github.com/nihcosnosaj/kindling)
![License](https://img.shields.io/github/license/nihcosnosaj/kindling)

### Why?
Writing standard `kind` configs can be very verbose. This cli script makes it all easy with just a single command!

### Installation
```bash
task build
sudo task install
```

### Usage
`kindling --workers 3 --name kindling-dev-labs`
