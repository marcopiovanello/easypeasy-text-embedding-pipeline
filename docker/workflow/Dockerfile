FROM registry.kokonatsu.cc/golang-tesseract-opencv:latest AS build

WORKDIR /usr/src/temporal
COPY . .

RUN GOAMD64=v3 CGO_ENABLED=1 go build -o workflow-trigger ./cmd/workflow/main.go

FROM alpine

RUN apk add --no-cache \
    opencv \
    tesseract-ocr \
    ca-certificates

WORKDIR /usr/bin
COPY --from=build /usr/src/temporal/workflow-trigger /usr/bin/workflow-trigger

ENTRYPOINT [ "/usr/bin/workflow-trigger"]