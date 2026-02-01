
  _  _____ _   _ _____  _     ___ _   _  ____ 
 | |/ /_ _| \ | |  _  \| |   |_ _| \ | |/ ___|
 | ' / | ||  \| | | |  | |    | ||  \| | |  _ 
 |  \  | || |\  | |_|  | |___ | || |\  | |_| |
 |_| \_\___|_| \_|____/|_____|___|_| \_|\____|

Kindling is a CLI tool I use for local testing and development with Kubernetes since I'm on a mac. My linux homelab machine(s) is/are still in development. :-) All it does is, based on CLI args, boots up an ephemeral k8s cluster locally using `kind` (Kubernetes in Docker)

[![CI](https://github.com/YOUR_USER/kindling/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USER/kindling/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USER/kindling)](https://goreportcard.com/report/github.com/YOUR_USER/kindling)

### Why?
Writing standard `kind` configs can be very verbose. This cli script makes it all easy with just a single command!

### Installation
```bash
task build
sudo task install
```