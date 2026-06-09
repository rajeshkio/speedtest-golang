# speedtest-golang

Application to run automated internet speed tests inside Docker containers and expose results as a JSON API.
Runs a speed test every 2 minutes using the official Ookla Speedtest CLI, stores results in MariaDB, and exposes them via a REST API.

# Architecture

Three containers:

- speedtest-db — MariaDB database storing all speed test results
- speedtest-collector — Go app running the Ookla speedtest CLI on a cron schedule and writing results to the DB
- speedtest-api — Go REST API reading from the DB and exposing results as JSON on port 8010

# Requirements

- Docker
- Docker Compose

# Deployment:

```
git clone https://github.com/rajeshkio/speedtest-golang.git
cd speedtest-golang
docker-compose up -d --build
```

# Usage

Access speed test results:

```
curl http://127.0.0.1:8010
```

Example response:

```
[{"ID":1,"TimeStamp":"2026-06-09T12:30:21Z","DownloadSpeed":"99","UploadSpeed":"98","Latency":"5.241","PublicIp":"157.20.184.37","ISP":"Vortex Infoway Private","Peers":"Pune Gazon Communications India Ltd India"}]
```

# Docker Hub

Pre-built images available for linux/amd64 and linux/arm64:

```
docker pull rk90229/speedtest-collector:v1.0.0
docker pull rk90229/speedtest-api:v1.0.0
```
