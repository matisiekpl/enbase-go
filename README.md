# Enbase ❤

[![Go Report Card](https://goreportcard.com/badge/github.com/enteam/enbase)](https://goreportcard.com/report/github.com/enteam/enbase)
[![Build Status](https://travis-ci.com/enteam/enbase.svg?branch=master)](https://travis-ci.com/enteam/enbase)
[![](https://img.shields.io/docker/pulls/enteam/enbase.svg)](https://hub.docker.com/r/enteam/enbase/)
[![](https://img.shields.io/docker/stars/enteam/enbase.svg)](https://hub.docker.com/r/enteam/enbase/)
[![](https://img.shields.io/github/license/enteam/enbase.svg)](https://github.com/enteam/enbase)
[![](https://img.shields.io/github/issues/enteam/enbase.svg)](https://github.com/enteam/enbase)

# Open Source NoSQL Realtime Database

## :star: Features
- [x] Fast & realtime
- [x] Servless
- [x] Powered by GoLang
- [x] Compatible with MongoDB and CosmosDB
- [x] Scalar horizontally
- [x] Powered by docker
- [x] Kubernetes compatible

## :rocket: Quick deployment
### Powered by docker compose :whale:
```
$ wget https://raw.githubusercontent.com/enteam/enbase/readme/docker-compose.yml
$ docker-compose up -d
```
### Powered by Kubernetes and Helm
```
$ helm repo add enbase https://enteam.github.io/enbase/charts
$ helm install enbase/enbase
```
