# ==========================================
# Build Stage
# ==========================================
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies (git is needed to clone the mtc library)
RUN apk add --no-cache go bash openssl build-base git

WORKDIR /app

# Clone the upstream MTC library — it is gitignored in this repo because it is
# an external dependency, so we fetch it fresh at image build time.
RUN git clone --depth=1 https://github.com/bwesterb/mtc.git mtc

# Copy the rest of the project (demo/, Dockerfile, etc.)
COPY . .

# Build the mtc-cli statically
RUN cd mtc && \
    CGO_ENABLED=0 go build -o ../mtc-cli ./cmd/mtc

# NOTE: setup.sh is NOT run at build time.
# Artifacts (CA, keys, MTC certs) are generated fresh at container startup
# via entrypoint.sh so they never expire due to a stale image.

# Build the Web Server statically
RUN cd demo && \
    CGO_ENABLED=0 go build -o website-server main.go

# ==========================================
# Runtime Stage
# ==========================================
# wolfi-base is used (instead of distroless) because setup.sh needs
# bash and openssl at runtime to generate fresh PKI artifacts on startup.
FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache bash openssl

# The Go backend expects to be run from the demo/ directory
# because it references relative paths like "static", "website.mtc", etc.
WORKDIR /app/demo

# Copy the CLI tool — main.go calls it via exec.Command("../mtc-cli")
COPY --from=builder /app/mtc-cli /app/mtc-cli

# Copy the server binary
COPY --from=builder /app/demo/website-server /app/demo/website-server

# Copy the setup script (runs at startup to generate fresh MTC artifacts)
COPY --from=builder /app/demo/setup.sh /app/demo/setup.sh

# Copy the static UI assets (these are not time-sensitive)
COPY --from=builder /app/demo/static /app/demo/static

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/demo/setup.sh

# Expose the HTTPS port
EXPOSE 8443

# On startup: generate fresh CA + MTC artifacts, then launch the server
ENTRYPOINT ["/app/entrypoint.sh"]
