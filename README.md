# GoBalancer

**GoBalancer** is a simple, pluggable, container-ready HTTP load balancer written in Go.  
It supports health checking, multiple load balancing strategies (e.g. Round Robin), and Docker-based deployments with sample backends.

---

## 🚀 Features

- ⚖️ Round-robin request distribution
  - more strategies coming _soon_
- 💚 Active health checks with retry logic
- 🔌 Plug-and-play load balancing strategies
- 🐳 Docker & Docker Compose support
- 📦 Lightweight, built with `distroless` for production

---

## 🪜Roadmap

- Configurable resilience strategies  
  Support customizable backoff, retry limits, and failure thresholds for health checks.

- Additional load balancing strategies  
  Implement and expose more strategies (e.g., random, least connections, weighted round-robin).

- Multiple server pools  
  Enable support for managing multiple pools concurrently, each with its own strategy (e.g., blue/green, service-based routing).

- Server status API  
  Endpoint (e.g., `/status`) that returns the current health of all registered servers.

---

## ⚙️ Configuration

GoBalancer reads its backend server list from a JSON file like this:
Applicable strategy will be selected based on the `strategy` value. Allowed strategies are: `rb` (more to come).
```json
{
  "serverPools": [
    {
      "name": "serverPool1",
      "strategy": "rb",
      "servers": [
        {
          "id": "1",
          "name": "Server 2137",
          "protocol": "http",
          "host": "localhost",
          "port": 2137,
          "healthcheckUrl": "/healthcheck"
        }
      ]
    }
  ]
}

```
See [servers.yaml](./servers.yaml) to learn yaml-based configuration.


## 🤖 Linting
This project uses [golangci-lint](https://golangci-lint.run/) for static code analysis.
Linting is enforced in CI, so pull requests will fail if linting issues are found.

To run linter locally:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./...
```

## 💻 Running locally
Simply run the following command to start GoBalancing:
```bash
go run ./cmd/main.go \
  --server-config=<path-to-servers.json-config> \
  --port=3000
```

Otherwise, you can run GoBalancer along with test apis (ports 2137 and 21370) in containers:
```bash
docker-compose up --build
```
