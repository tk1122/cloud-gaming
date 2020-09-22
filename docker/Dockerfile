FROM golang:1.15 AS builder
RUN apt-get update && apt-get install -y \
    pkg-config \
    libvpx-dev
WORKDIR /app/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY . .
RUN go build -o main cmd/worker/*
CMD ["./main"]
