prepare:
	go get -d -v -t ./...
	docker-compose up -d

test: prepare
	go test -short ./...

update:
	go get -u -v -d -t ./...

.PHONY: test
