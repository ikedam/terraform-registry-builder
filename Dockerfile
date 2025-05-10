FROM golang:1.23.4-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o terraform-registry-builder .

FROM alpine:3.21.3

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/terraform-registry-builder /terraform-registry-builder

ENTRYPOINT ["/terraform-registry-builder"]
