PREFIX := /usr/local
VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

TELEGRAF := telegraf$(shell go tool dist env | grep -q 'GOOS=.windows.' && echo .exe)

LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)
ifdef VERSION
	LDFLAGS += -X main.version=$(VERSION)
endif

all:
	$(MAKE) deps
	$(MAKE) telegraf

deps:
	go get github.com/sparrc/gdm
	gdm restore

telegraf:
	go build -i -o $(TELEGRAF) -ldflags "$(LDFLAGS)" ./cmd/telegraf/telegraf.go

go-install:
	go install -ldflags "-w -s $(LDFLAGS)" ./cmd/telegraf

install: telegraf
	mkdir -p $(DESTDIR)$(PREFIX)/bin/
	cp $(TELEGRAF) $(DESTDIR)$(PREFIX)/bin/

test:
	go test -short ./...

test-windows:
	go test ./plugins/inputs/ping/...
	go test ./plugins/inputs/win_perf_counters/...
	go test ./plugins/inputs/win_services/...

lint:
	go vet ./...

test-all: lint
	go test ./...

package:
	./scripts/build.py --package --platform=all --arch=all

clean:
	-rm -f telegraf
	-rm -f telegraf.exe

docker-image:
	./scripts/build.py --package --platform=linux --arch=amd64
	cp build/telegraf*$(COMMIT)*.deb .
	docker build -f scripts/dev.docker --build-arg "package=telegraf*$(COMMIT)*.deb" -t "telegraf-dev:$(COMMIT)" .

# Run all docker containers necessary for integration tests
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
	docker run --name cratedb \
		-p "6543:5432" \
		-d crate crate \
		-Cnetwork.host=0.0.0.0 \
		-Ctransport.host=localhost \
		-Clicense.enterprise=false

# Run docker containers necessary for integration tests; skipping services provided
# by CircleCI
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
	docker run --name cratedb \
		-p "6543:5432" \
		-d crate crate \
		-Cnetwork.host=0.0.0.0 \
		-Ctransport.host=localhost \
		-Clicense.enterprise=false

docker-kill:
	-docker kill aerospike elasticsearch kafka memcached mqtt mysql nats nsq \
		openldap postgres rabbitmq redis riemann zookeeper cratedb
	-docker rm aerospike elasticsearch kafka memcached mqtt mysql nats nsq \
		openldap postgres rabbitmq redis riemann zookeeper cratedb

.PHONY: deps telegraf telegraf.exe install test test-windows lint test-all \
	package clean docker-run docker-run-circle docker-kill docker-image
