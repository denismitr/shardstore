FROM golang:latest as builder

WORKDIR /build

COPY go.mod go.sum ./
COPY vendor/ ./vendor

COPY cmd/ ./cmd
COPY internal/ ./internal
COPY pkg/ ./pkg

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o filestore ./cmd/filestore/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /source

COPY --from=builder /build/filestore .

RUN chmod +x ./filestore

CMD ["./filestore"]