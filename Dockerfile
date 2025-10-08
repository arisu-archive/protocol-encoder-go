# Builder stage
FROM golang:1.25-alpine AS builder

# Configure arguments
ARG JAPAN_OFFSET
ARG GLOBAL_OFFSET

# Set the working directory inside the container
WORKDIR /app

# We need to install git to fetch dependencies
RUN apk add --no-cache git cmake samurai build-base pkgconf linux-headers curl

# Copy the entire project into the container
COPY . .

# Download dependencies
RUN go mod download

# Apply patch if exists
RUN chmod +x ./scripts/apply_patch.sh
RUN ./scripts/apply_patch.sh

# Configure & build Unicorn
RUN cmake -S unicorn -B build/unicorn -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DUNICORN_BUILD_SHARED=ON \
    -DUNICORN_ARCH="aarch64" && \
    cmake --build build/unicorn --config Release -j$(nproc)

# Setup CGO environment variables
ENV CGO_CFLAGS="-I/app/build/unicorn/include" \
    CGO_LDFLAGS="-L/app/build/unicorn -lunicorn" \
    CGO_ENABLED=1

# Build the Go application
RUN go build -v -trimpath -ldflags="-s -w" -o protocol-encoder ./main.go

# Download the library dependencies
RUN chmod +x ./scripts/download_library.sh
RUN ./scripts/download_library.sh "com.YostarJP.BlueArchive" "libraries/japan/libil2cpp.so"
RUN ./scripts/download_library.sh "com.nexon.bluearchive" "libraries/global/libil2cpp.so"

# Runtime stage
FROM alpine:latest

WORKDIR /root/app

COPY --from=builder /app/protocol-encoder .
COPY --from=builder /app/libraries/japan/libil2cpp.so ./libraries/japan/libil2cpp.so
COPY --from=builder /app/libraries/global/libil2cpp.so ./libraries/global/libil2cpp.so
COPY --from=builder /app/config.example.yaml ./config.yaml
COPY --from=builder /app/build/unicorn/libunicorn.so* /usr/local/lib/

# Update the dynamic linker cache
RUN ldconfig /usr/local/lib

EXPOSE 8080

ENTRYPOINT ["./protocol-encoder", "serve"]