.PHONY: all build run demo test bench k6 docker-up docker-down clean

# Go Application targets
build:
	go build -ldflags="-w -s" -o bin/ratelimiter cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v -race ./tests/...

test-bench:
	go test -bench=. -benchmem ./tests/...

# Instant Run (Zero Setup via Python)
demo:
	python run_demo.py

# High-Throughput Benchmarks
bench:
	python benchmarks/benchmark.py -n 10000 -c 50

simulate:
	python benchmarks/simulate_traffic.py

k6:
	docker compose -f deployments/docker-compose.yml run --rm k6

# Docker Infrastructure (Rate Limiter + Redis + Prometheus + Grafana)
docker-up:
	docker compose -f deployments/docker-compose.yml up --build -d
	@echo "=========================================================="
	@echo " System Initialized Successfully!"
	@echo " Rate Limiter REST:   http://localhost:8080"
	@echo " Rate Limiter gRPC:   localhost:50051"
	@echo " Web Visualizer UI:   http://localhost:8080/dashboard"
	@echo " Prometheus Metrics:  http://localhost:9090"
	@echo " Grafana Dashboards:  http://localhost:3000 (admin/admin)"
	@echo "=========================================================="

docker-down:
	docker compose -f deployments/docker-compose.yml down -v

clean:
	rm -rf bin/
