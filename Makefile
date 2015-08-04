prepare:
	go get -d -v -t ./...
	docker-compose up -d --no-recreate

test-short: prepare
	go test -short ./...

test: prepare
	go test ./...

update:
	go get -u -v -d -t ./...

.PHONY: test
