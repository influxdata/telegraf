static:
	@echo "Building static linux binary..."
	@CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" ./cmd/telegraf

# Input plugins (Please keep alphabetized)
plugin-logparser:
	docker-compose -f plugins/inputs/logparser/test/docker-compose.yml up

plugin-mysql:
	docker-compose -f plugins/inputs/mysql/test/docker-compose.yml up
