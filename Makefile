VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

LDFLAGS := -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
	LDFLAGS += -X main.version=$(VERSION)
endif


# Standard Telegraf build
default: prepare build

# Windows build
windows: prepare-windows build-windows

# Only run the build (no dependency grabbing)
build:
	go install -ldflags "$(LDFLAGS)" ./...

build-windows:
	GOOS=windows GOARCH=amd64 go build -o telegraf.exe \
		 -ldflags "$(LDFLAGS)" \
		./cmd/telegraf/telegraf.go

build-for-docker:
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o telegraf \
		 -ldflags "$(LDFLAGS)" \
		./cmd/telegraf/telegraf.go

# run package script
package:
	./scripts/build.py --package --version="$(VERSION)" --platform=linux --arch=all --upload

# Get dependencies and use gdm to checkout changesets
prepare:
	go get github.com/sparrc/gdm
	gdm restore

# Use the windows godeps file to prepare dependencies
prepare-windows:
	go get github.com/sparrc/gdm
	gdm restore
	gdm restore -f Godeps_windows

# Run all docker containers necessary for unit tests
docker-run:
	docker run --name aerospike -p "3000:3000" -d aerospike/aerospike-server:3.9.0
	docker run --name zookeeper -p "2181:2181" -d wurstmeister/zookeeper
	docker run --name kafka \
		--link zookeeper:zookeeper \
		-e KAFKA_ADVERTISED_HOST_NAME=localhost \
		-e KAFKA_ADVERTISED_PORT=9092 \
		-e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
		-e KAFKA_CREATE_TOPICS="test:1:1" \
		-p "9092:9092" \
		-d wurstmeister/kafka
	docker run --name elasticsearch -p "9200:9200" -p "9300:9300" -d elasticsearch:5
	docker run --name mysql -p "3306:3306" -e MYSQL_ALLOW_EMPTY_PASSWORD=yes -d mysql
	docker run --name memcached -p "11211:11211" -d memcached
	docker run --name postgres -p "5432:5432" -d postgres
	docker run --name rabbitmq -p "15672:15672" -p "5672:5672" -d rabbitmq:3-management
	docker run --name redis -p "6379:6379" -d redis
	docker run --name nsq -p "4150:4150" -d nsqio/nsq /nsqd
	docker run --name mqtt -p "1883:1883" -d ncarlier/mqtt
	docker run --name riemann -p "5555:5555" -d stealthly/docker-riemann
	docker run --name nats -p "4222:4222" -d nats
	docker run --name openldap \
		-e SLAPD_CONFIG_ROOTDN="cn=manager,cn=config" \
		-e SLAPD_CONFIG_ROOTPW="secret" \
		-p "389:389" -p "636:636" \
		-d cobaugh/openldap-alpine

# Run docker containers necessary for CircleCI unit tests
docker-run-circle:
	docker run --name aerospike -p "3000:3000" -d aerospike/aerospike-server:3.9.0
	docker run --name zookeeper -p "2181:2181" -d wurstmeister/zookeeper
	docker run --name kafka \
		--link zookeeper:zookeeper \
		-e KAFKA_ADVERTISED_HOST_NAME=localhost \
		-e KAFKA_ADVERTISED_PORT=9092 \
		-e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
		-e KAFKA_CREATE_TOPICS="test:1:1" \
		-p "9092:9092" \
		-d wurstmeister/kafka
	docker run --name elasticsearch -p "9200:9200" -p "9300:9300" -d elasticsearch:5
	docker run --name nsq -p "4150:4150" -d nsqio/nsq /nsqd
	docker run --name mqtt -p "1883:1883" -d ncarlier/mqtt
	docker run --name riemann -p "5555:5555" -d stealthly/docker-riemann
	docker run --name nats -p "4222:4222" -d nats
	docker run --name openldap \
		-e SLAPD_CONFIG_ROOTDN="cn=manager,cn=config" \
		-e SLAPD_CONFIG_ROOTPW="secret" \
		-p "389:389" -p "636:636" \
		-d cobaugh/openldap-alpine

# Kill all docker containers, ignore errors
docker-kill:
	-docker kill nsq aerospike redis rabbitmq postgres memcached mysql zookeeper kafka mqtt riemann nats elasticsearch openldap
	-docker rm nsq aerospike redis rabbitmq postgres memcached mysql zookeeper kafka mqtt riemann nats elasticsearch openldap

# Run full unit tests using docker containers (includes setup and teardown)
test: vet docker-kill docker-run
	# Sleeping for kafka leadership election, TSDB setup, etc.
	sleep 60
	# SUCCESS, running tests
	go test -race ./...

# Run "short" unit tests
test-short: vet
	go test -short ./...

# Run windows specific tests
test-windows: vet
	go test ./plugins/inputs/ping/...
	go test ./plugins/inputs/win_perf_counters/...

vet:
	go vet ./...

.PHONY: test test-short vet build default
