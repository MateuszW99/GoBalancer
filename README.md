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


## Planned Features

- Configurable resilience strategies  
  Support customizable backoff, retry limits, and failure thresholds for health checks.

- GitHub Actions CI  
  Add automated testing, linting, and build checks using GitHub Actions for every push/PR.

- Additional load balancing strategies  
  Implement and expose more strategies (e.g., random, least connections, weighted round-robin).

- Multiple server pools  
  Enable support for managing multiple pools concurrently, each with its own strategy (e.g., blue/green, service-based routing).

- Server status API  
  Endpoint (e.g., `/status`) that returns the current health of all registered servers.

- YAML configuration support  
  Allow loading config from YAML in addition to JSON for improved readability and flexibility.


---

## ⚙️ Configuration

GoBalancer reads its backend server list from a JSON file like this:

```json
[
  {
    "ID": "1",
    "Name": "Server 1",
    "Protocol": "http",
    "Host": "testapi",
    "Port": 2137,
    "Url": "http://localhost:2137",
    "HealthcheckUrl": "/healthcheck"
  }
]
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
