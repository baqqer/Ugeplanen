# Stage 1: Build the statically compiled Go binary
FROM golang:alpine AS builder

# Set build configurations
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Copy go.mod and source files
COPY go.mod ./
COPY main.go ./

# Compile binary and strip debug symbols to minimize file size
RUN go build -ldflags="-w -s" -o ugeplanen main.go

# Stage 2: Create highly-optimized final runner image
FROM alpine:3.21

# Create a non-root system user and group for security
RUN addgroup -S ugeplanen && adduser -S ugeplanen -G ugeplanen

WORKDIR /app

# Copy the compiled binary and static templates
COPY --from=builder /build/ugeplanen ./
COPY templates/ ./templates/
COPY static/ ./static/

# Set correct ownership on the app directory so the app can create plan.json
RUN chown -R ugeplanen:ugeplanen /app

# Switch to the non-root user
USER ugeplanen

# Expose default application port
EXPOSE 9000

# Run the binary
ENTRYPOINT ["/app/ugeplanen"]
