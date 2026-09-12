images:
	docker buildx build --file ./docker/ocr/Dockerfile -t registry.kokonatsu.cc/epocr-ocr-worker:latest --push .
	docker buildx build --file ./docker/convert/Dockerfile -t registry.kokonatsu.cc/epocr-convert-worker:latest --push .
	docker buildx build --file ./docker/embedding/Dockerfile -t registry.kokonatsu.cc/epocr-embedding-worker:latest --push .
	docker buildx build --file ./docker/workflow/Dockerfile -t registry.kokonatsu.cc/epocr-workflow-trigger:latest --push .
	docker buildx build --file ./docker/upload/Dockerfile -t registry.kokonatsu.cc/epocr-upload-service:latest --push .

builder:
	docker buildx build --file ./docker/Dockerfile.builder -t registry.kokonatsu.cc/golang-tesseract-opencv:latest --push .