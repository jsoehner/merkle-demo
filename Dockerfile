# ==========================================
# Build Stage
# ==========================================
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies
RUN apk add --no-cache go bash openssl build-base git

WORKDIR /app

# ── 1. Clone the upstream MTC library (gitignored)
RUN git clone --depth=1 https://github.com/bwesterb/mtc.git mtc

# ── 2. Clone the DigiCert ca-extension-mtc-playground
RUN git clone --depth=1 https://github.com/digicert/ca-extension-mtc-playground.git ca-extension-mtc-playground

# Copy the rest of the project
COPY . .

# ── 3. Build mtc-cli (bwesterb/mtc)
RUN cd mtc && \
    CGO_ENABLED=0 go build -o ../mtc-cli ./cmd/mtc

# ── 4. Build the MTC demo website server
RUN cd demo && \
    CGO_ENABLED=0 go build -o website-server main.go

# ── 5. Build the playground server
RUN cd playground && \
    CGO_ENABLED=0 go build -o playground-server main.go

# ── 6. Build DigiCert playground standalone tools (no DB/network needed)
RUN cd ca-extension-mtc-playground && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/demo-embedded-cert ./cmd/demo-embedded-cert/ && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/mtc-verify-cert    ./cmd/mtc-verify-cert/ && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/mtc-conformance    ./cmd/mtc-conformance/ && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/mtc-interop        ./cmd/mtc-interop/

# ==========================================
# Runtime Stage
# ==========================================
FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache bash openssl

# MTC demo website workdir
WORKDIR /app/demo

# ── Core binaries
COPY --from=builder /app/mtc-cli        /app/mtc-cli
COPY --from=builder /app/demo/website-server /app/demo/website-server

# ── Demo website setup scripts + static assets
COPY --from=builder /app/demo/setup.sh  /app/demo/setup.sh
COPY --from=builder /app/demo/static    /app/demo/static

# ── Playground server + static assets
COPY --from=builder /app/playground/playground-server /app/playground/playground-server
COPY --from=builder /app/playground/static            /app/playground/static

# ── DigiCert playground standalone tools
COPY --from=builder /tmp/demo-embedded-cert /usr/local/bin/demo-embedded-cert
COPY --from=builder /tmp/mtc-verify-cert    /usr/local/bin/mtc-verify-cert
COPY --from=builder /tmp/mtc-conformance    /usr/local/bin/mtc-conformance
COPY --from=builder /tmp/mtc-interop        /usr/local/bin/mtc-interop

# ── Entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/demo/setup.sh \
             /app/demo/website-server \
             /app/playground/playground-server \
             /usr/local/bin/demo-embedded-cert \
             /usr/local/bin/mtc-verify-cert \
             /usr/local/bin/mtc-conformance \
             /usr/local/bin/mtc-interop

# Expose MTC Demo (HTTPS) + Playground (HTTP)
EXPOSE 8443 8444

# Health check — playground starts last, so if it's up both servers are ready
HEALTHCHECK --interval=5s --timeout=3s --start-period=45s --retries=3 \
  CMD bash -c 'echo > /dev/tcp/localhost/8444' || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
