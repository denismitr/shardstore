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
up:
	docker-compose -f docker-compose-local.yml up -d --build

.PHONY: down
down:
	docker-compose -f docker-compose-local.yml down --remove-orphans

.PHONY: docker-remove
docker-remove:
	docker rm --force `docker ps -a -q` || true