<p align="center"><img width="30%" src="images/logo.png"/></p>

<h1 align="center">Enbase ⚡️</h1>

<p align="center">
  
<a href="https://goreportcard.com/report/github.com/enteam/enbase">
  <img src="https://goreportcard.com/badge/github.com/enteam/enbase">
</a>

<a href="https://travis-ci.com/enteam/enbase">
  <img src="https://travis-ci.com/enteam/enbase.svg?branch=master">
</a>

<a href="https://hub.docker.com/r/enteam/enbase/">
  <img src="https://img.shields.io/docker/pulls/enteam/enbase.svg">
</a>

<a href="https://hub.docker.com/r/enteam/enbase/">
  <img src="https://img.shields.io/docker/stars/enteam/enbase.svg">
</a>

<a href="https://github.com/enteam/enbase">
  <img src="https://img.shields.io/github/license/enteam/enbase.svg">
</a>

<a href="https://github.com/enteam/enbase">
  <img src="https://img.shields.io/github/issues/enteam/enbase.svg">
</a>

</p>

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
$ wget https://raw.githubusercontent.com/enteam/enbase/master/docker-compose.yml
$ docker-compose up -d
```
### Powered by Kubernetes and Helm
```
$ helm repo add enbase https://enteam.github.io/enbase/charts
$ helm install enbase/enbase
```
