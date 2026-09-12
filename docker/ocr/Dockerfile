FROM registry.kokonatsu.cc/golang-tesseract-opencv:latest AS build

WORKDIR /usr/src/temporal
COPY . .

RUN GOAMD64=v3 CGO_ENABLED=1 go build -o ocr-worker ./cmd/ocr/main.go

FROM alpine

RUN apk update && apk add --no-cache \
    opencv \
    opencv-dev \
    tesseract-ocr \
    tesseract-ocr-dev \
    tesseract-ocr-data-ita \
    tesseract-ocr-data-eng \
    ca-certificates

WORKDIR /usr/bin
COPY --from=build /usr/src/temporal/ocr-worker /usr/bin/ocr-worker

ENTRYPOINT [ "/usr/bin/ocr-worker"]