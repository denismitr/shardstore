PROTO_OUT_DIR=./pkg/storeserver/v1
PROTO_SOURCE_DIR=./api/storeserver/v1

.PHONY: proto
proto:
	$(info making directory ${PROTO_OUT_DIR})
	@mkdir -p ${PROTO_OUT_DIR}
	$(info compiling protoc files into ${PROTO_SOURCE_DIR})
	protoc \
		-I ${PROTO_SOURCE_DIR} \
		--include_imports \
		--go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
        --go-grpc_out=$(PROTO_OUT_DIR)  --go-grpc_opt=paths=source_relative \
		--descriptor_set_out=$(PROTO_OUT_DIR)/api.pb \
		./${PROTO_SOURCE_DIR}/*.proto

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: build
build:
	go build -o bin/filegateway cmd/filegateway/main.go
	go build -o bin/filestore cmd/filestore/main.go

.PHONY: up
up: deps
	docker-compose -f docker-compose-local.yaml up -d --build

.PHONY: down
down:
	docker-compose -f docker-compose-local.yaml down --remove-orphans

.PHONY: docker-remove
docker-remove:
	docker rm --force `docker ps -a -q` || true

.PHONY: test
test:
	go test -v ./...

.PHONY: upload
upload:
	curl -X PUT -F 'file=@./samples/1.png' http://localhost:8080/upload