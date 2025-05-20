# News Service

A simple blogging/news service built in Go, using MongoDB and server‑side HTMX rendering templates. It covers:

* Clean application layering (handler → service → repository)
* MongoDB integration with BSON marshalling
* Unit and integration tests (including Docker‑backed MongoDB)
* Docker & Docker Compose for easy local development
* Makefile for common tasks (build, test, lint, run, compose)

---

## Table of Contents

* [Prerequisites](#prerequisites)
* [Environment Variables](#environment-variables)
* [Local Development](#local-development)

  * [Build & Run](#build--run)
  * [Testing](#testing)
* [Docker](#docker)

  * [Build Image](#build-image)
  * [Run Container](#run-container)
* [Docker Compose](#docker-compose)

---

## Prerequisites

* Go **1.24**
* Docker & Docker Compose (for containerized workflows)
* MongoDB (if running locally without Docker)

---

## Environment Variables

Create a `.env` file in the project root, or export these in your shell:

```bash
MONGO_USER=admin
MONGO_PASSWORD=secretpassword
MONGO_HOST=localhost      # or 'mongo' when using Docker Compose
MONGO_PORT=27017
MONGO_NAME=news_db

SERVER_PORT=8080
SERVER_IS_DEV=true        # 'true' enables debug logging
```

---

## Local Development

### Build & Run

```bash
# Build the binary
make build

# Run locally (uses env variables)
make run

# Run in a docker container
make compose-up
```

You should see log output indicating MongoDB connection and server start. Visit `http://localhost:8080/ping` to confirm.

### Testing

* **Unit Tests**: run all unit tests

```bash
go test ./... -short
```

- **Integration Tests**: requires Docker

```bash
make compose-up   # starts MongoDB in Docker
go test ./internal/storage/mongo/post -timeout 2m
make compose-down
````

---

## Docker

### Build Image

```bash
make docker-build
```

### Run Container

```bash
make docker-run
```

This uses the image built above, linking it to a running MongoDB (set via environment variables).

---

## Docker Compose

Bring up the full stack (service + MongoDB):

```bash
make compose-up
```

Once started, use:

* HTTP API: `http://localhost:${SERVER_PORT}`
* Mongo shell on `mongodb://localhost:27017`

To tear down:

```bash
make compose-down
```