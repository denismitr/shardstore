FROM golang:latest as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY cmd/ ./cmd
COPY internal/ ./internal

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o -o filestore ./cmd/filestore/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /source

COPY --from=builder /build/filestore .

RUN chmod +x ./filestore

CMD ["./filestore"]