FROM registry.kokonatsu.cc/golang-tesseract-opencv:latest AS build

WORKDIR /usr/src/temporal
COPY . .

RUN GOAMD64=v3 CGO_ENABLED=1 go build -o embedding-worker ./cmd/embedding/main.go

FROM alpine

RUN apk add --no-cache \
    opencv \
    opencv-dev \
    tesseract-ocr \
    tesseract-ocr-dev \
    ca-certificates

WORKDIR /usr/bin
COPY --from=build /usr/src/temporal/embedding-worker /usr/bin/embedding-worker

ENTRYPOINT [ "/usr/bin/embedding-worker"]