# syntax=docker/dockerfile:1
# Multi-stage build: compile with Go, run a tiny static binary.
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /app .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app .
USER nobody
ENTRYPOINT ["./app"]
