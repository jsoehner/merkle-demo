# ==========================================
# Build Stage
# ==========================================
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies
RUN apk add --no-cache go bash openssl build-base

WORKDIR /app

# Copy the entire project
COPY . .

# Build the mtc-cli statically
RUN cd mtc && \
    CGO_ENABLED=0 go build -o ../mtc-cli ./cmd/mtc

# Run the setup script to generate the CA, Keys, and MTC artifacts
# We run this during the build so the artifacts are baked into the container
RUN cd demo && \
    chmod +x setup.sh && \
    ./setup.sh

# Build the Web Server statically
RUN cd demo && \
    CGO_ENABLED=0 go build -o website-server main.go

# ==========================================
# Runtime Stage (Distroless)
# ==========================================
FROM cgr.dev/chainguard/static:latest

# The Go backend expects to be run from the demo/ directory 
# because it references files like "static" and "website.mtc"
WORKDIR /app/demo

# Copy the CLI tool to the parent directory as expected by main.go (exec.Command("../mtc-cli"))
COPY --from=builder /app/mtc-cli /app/mtc-cli

# Copy the server binary
COPY --from=builder /app/demo/website-server /app/demo/website-server

# Copy all the generated PKI and website files
COPY --from=builder /app/demo/ca /app/demo/ca
COPY --from=builder /app/demo/website.mtc /app/demo/website.mtc
COPY --from=builder /app/demo/website.vw /app/demo/website.vw
COPY --from=builder /app/demo/website.pem /app/demo/website.pem
COPY --from=builder /app/demo/website.key /app/demo/website.key
COPY --from=builder /app/demo/static /app/demo/static

# Expose the HTTPS port
EXPOSE 8443

# Start the server
ENTRYPOINT ["/app/demo/website-server"]
