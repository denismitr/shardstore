FROM golang:latest as builder

WORKDIR /build

COPY go.mod ./
COPY vendor/ ./vendor

COPY cmd/ ./cmd
COPY internal/ ./internal
COPY pkg/ ./pkg

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o filegateway ./cmd/filegateway/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /source

COPY --from=builder /build/filegateway .

EXPOSE 8080

RUN chmod +x ./filegateway

CMD ["./filegateway"]