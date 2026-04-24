# Build Stage
FROM golang:1.26-alpine AS build
WORKDIR /source

# Copy source code
COPY . .

# Build the application
RUN go build -o wiz-controller ./cmd/wiz-controller

# Runtime Stage
FROM alpine:latest
WORKDIR /app

# Install ca-certificates for any HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy binary from build stage
COPY --from=build /source/wiz-controller .

# Copy config and static files
COPY config.json .
COPY static ./static

# The app expects port 80
EXPOSE 80

ENTRYPOINT ["./wiz-controller"]

LABEL org.opencontainers.image.description="A simple WiZ light controller for local networks"
