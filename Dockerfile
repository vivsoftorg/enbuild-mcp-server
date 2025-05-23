# Base stage for building binaries (using Go)
# -------------------------------------------
FROM golang:1.23 AS devbuild
WORKDIR /build

# Cache mod downloads and build artifacts
RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY . ./

# Build the server with minimal executable size
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w \
    -X mcp-server-enbuild/version.GitCommit=$(git rev-parse HEAD) \
    -X mcp-server-enbuild/version.BuildDate=$(git show --no-show-signature -s --format=%cd --date=format:'%Y-%m-%dT%H:%M:%SZ' HEAD)" \
    -o mcp-server-enbuild /build/

# Stage for development usage
# -------------------------------------------
FROM alpine:3.21 AS dev
WORKDIR /server

# Add a non-root user to enhance security
RUN addgroup -S mygroup && adduser -S myuser -G mygroup
USER myuser

# Copy the binary from the build stage
COPY --from=devbuild /build/mcp-server-enbuild .

# Command to run the server
CMD ["./mcp-server-enbuild", "stdio"]

# Release stage for production usage using CI-built binaries
# -------------------------------------------
FROM alpine:3.21 AS release-default

# Export BIN_NAME for the CMD below
ARG BIN_NAME=mcp-server-enbuild
ARG PRODUCT_VERSION
ARG PRODUCT_REVISION

# Target architecture and OS (default to current platform if not specified)
ARG TARGETOS=linux
ARG TARGETARCH=amd64

LABEL version="${PRODUCT_VERSION}"
LABEL revision="${PRODUCT_REVISION}"

# Use default value for BIN_NAME in case it's not provided
COPY dist/${TARGETOS}/${TARGETARCH}/${BIN_NAME} /bin/mcp-server-enbuild

CMD ["/bin/mcp-server-enbuild", "stdio"]

# Default target is 'dev'
FROM dev
