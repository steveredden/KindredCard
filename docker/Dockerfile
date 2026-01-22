# Stage 1: Build CSS with Node.js
FROM node:20-alpine AS css-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install Node dependencies
RUN npm install

# Copy CSS source and templates (needed for Tailwind content scanning)
COPY web/templates ./web/templates
COPY web/static/js ./web/static/js
COPY web/static/css/input.css ./web/static/css/
COPY tailwind.config.js ./

# Build production CSS (minified, purged)
RUN npm run build:css:prod

# Stage 2: Build Go binary
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS go-builder

ARG TARGETARCH
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy built CSS from previous stage
COPY --from=css-builder /app/web/static/css/output.css ./web/static/css/

# Build the application
ARG VERSION=v0.0.0-dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -ldflags="-s -w -X 'main.ReleaseVersion=${VERSION}'" -o kindredcard cmd/kindredcard/main.go

# Stage 3: Final runtime image
FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/steveredden/KindredCard"
LABEL org.opencontainers.image.licenses="AGPL-3.0-only"
LABEL org.opencontainers.image.description="Personal CRM for managing contacts and relationships"

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/kindredcard .

# Copy web assets (templates, JS, built CSS)
COPY --from=go-builder /app/web ./web

# Create non-root user
RUN addgroup -g 1000 kindredcard && \
    adduser -D -u 1000 -G kindredcard kindredcard && \
    chown -R kindredcard:kindredcard /app

USER kindredcard

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=60s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./kindredcard"]