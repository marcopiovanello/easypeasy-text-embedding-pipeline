FROM registry.kokonatsu.cc/golang-tesseract-opencv:latest AS build

WORKDIR /usr/src/temporal
COPY . .

ENV CGO_ENABLED=1

RUN GOAMD64=v3 go build -o convert-worker ./cmd/convert/main.go

FROM alpine

RUN apk add --no-cache \
    opencv \
    opencv-dev \
    tesseract-ocr \
    tesseract-ocr-dev \
    ca-certificates

WORKDIR /usr/bin
COPY --from=build /usr/src/temporal/convert-worker /usr/bin/convert-worker

ENTRYPOINT [ "/usr/bin/convert-worker" ]