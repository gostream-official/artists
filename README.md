# artists

---

*artists* is part of the *gostream* project. *gostream* is simple music database. *artists* is a service for artist management.

Features:

- query artists
- create artists
- update artists
- delete artists

---

## Quickstart

For a quick start with *gostream*, use the official deployment repository: [deployment](https://github.com/gostream-official/deployment)

For a quick start with *artists*, use the official docker container:

```sh
$ docker pull ghcr.io/gostream-official/artists:latest
```

or start with a docker-compose file:

```yml
version: '3.8'

services:

  artists:
    image: ghcr.io/gostream-official/artists:latest
    container_name: artists
    environment:
      MONGO_USERNAME: root
      MONGO_PASSWORD: example
      MONGO_HOST: mongo:27017
    ports:
      - "9871:9871"

  mongo:
    image: mongo:latest
    container_name: mongo
    ports:
      - 27017:27017
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: example
```

## Setup

To get *artists* up and running, follow the instructions below.

### Platforms

Officially supported development platforms are:

- Windows
- MacOS
- Linux

### Go

The *artists* project is written in *Go*, hence it is required to install *Go*. For the latest version of *Go*, check: https://go.dev/doc/install

## Build and Run

Build the *artists* project using:

```sh
$ go build -o bin/artists cmd/main.go
```

Run the *artists* project using:

```sh
$ MONGO_USERNAME=root MONGO_PASSWORD=example go run cmd/main.go
```

## Debugging

Debug the *artists* project using the provided `launch.json` file for *Visual Studio Code*.

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/main.go",
      "showLog": true,
      "internalConsoleOptions": "openOnSessionStart",
      "env": {
        "MONGO_USERNAME": "root",
        "MONGO_PASSWORD": "example",
        "MONGO_HOST": "127.0.0.1:27017"
      }
    }
  ]
}
```