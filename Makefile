reqs ?= 500

start-pp:
	docker compose -f payment-processor/docker-compose.yml up -d

dev:
	docker compose -f payment-processor/docker-compose.yml up -d
	COMPOSE_BAKE=true docker compose -f docker-compose.dev.yml up

devb:
	docker compose -f payment-processor/docker-compose.yml up -d
	COMPOSE_BAKE=true docker compose -f docker-compose.dev.yml up --build

compose-down:
	docker compose -f docker-compose.dev.yml down
	docker compose -f payment-processor/docker-compose.yml down

docker-clean:
	docker image prune -f
	docker image rm rinha-2025-api1
	docker image rm rinha-2025-api2

down: compose-down docker-clean

rs: compose-down docker-clean devb

# K6

ci:
	K6_WEB_DASHBOARD_OPEN=false K6_WEB_DASHBOARD=true K6_WEB_DASHBOARD_EXPORT='report.html' K6_WEB_DASHBOARD_PERIOD=2s k6 run -e MAX_REQUESTS=$(reqs) rinha-test/rinha.js

ci-m:
	K6_WEB_DASHBOARD_OPEN=false K6_WEB_DASHBOARD=true K6_WEB_DASHBOARD_EXPORT='report.html' K6_WEB_DASHBOARD_PERIOD=2s k6 run -e MAX_REQUESTS=4000 rinha-test/rinha.js

check%:
	node scripts/check-uuid-logs.js logs/api$*/app.jsonl