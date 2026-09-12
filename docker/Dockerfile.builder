FROM golang:1-alpine

RUN apk add --no-cache \
    tesseract-ocr-dev \
    opencv-dev \
    build-base \
    pkgconfig \
    gcc \
    g++ \
    musl-dev

RUN mkdir -p /usr/lib/pkgconfig && \
    echo 'prefix=/usr' > /usr/lib/pkgconfig/opencv4.pc && \
    echo 'exec_prefix=${prefix}' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'libdir=${prefix}/lib' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'includedir=${prefix}/include/opencv4' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo '' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'Name: OpenCV' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'Description: Open Source Computer Vision Library' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'Version: 4.x' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'Libs: -L${libdir} -lopencv_core -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_videoio -lopencv_calib3d -lopencv_features2d -lopencv_objdetect -lopencv_dnn -lopencv_video -lopencv_photo' >> /usr/lib/pkgconfig/opencv4.pc && \
    echo 'Cflags: -I${includedir}' >> /usr/lib/pkgconfig/opencv4.pc