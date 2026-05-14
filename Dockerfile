# ==========================================
# Build Stage
# ==========================================
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies
RUN apk add --no-cache go bash openssl build-base git

WORKDIR /app

# ── 1. Clone and build the upstream MTC library (bwesterb/mtc)
RUN git clone --depth=1 https://github.com/bwesterb/mtc.git mtc && \
    cd mtc && \
    go mod download && \
    CGO_ENABLED=0 go build -v -o /app/mtc-cli ./cmd/mtc

# ── 2. Clone and build DigiCert playground standalone tools
RUN git clone --depth=1 https://github.com/digicert/ca-extension-mtc-playground.git ca-extension-mtc-playground && \
    cd ca-extension-mtc-playground && \
    go mod download && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/demo-embedded-cert ./cmd/demo-embedded-cert/ && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/mtc-verify-cert    ./cmd/mtc-verify-cert/ && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/mtc-conformance    ./cmd/mtc-conformance/ && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/mtc-interop        ./cmd/mtc-interop/

# ── 3. Copy the rest of the project
COPY . .

# ── 4. Build the MTC demo website server
RUN cd demo && \
    CGO_ENABLED=0 go build -o website-server main.go

# ── 5. Build the playground server
RUN cd playground && \
    CGO_ENABLED=0 go build -o playground-server main.go

# ==========================================
# Common Runtime Base
# ==========================================
FROM cgr.dev/chainguard/wolfi-base AS runtime-base
RUN apk add --no-cache bash openssl bc curl
WORKDIR /app

# ==========================================
# Demo Runtime Stage
# ==========================================
FROM runtime-base AS demo-runtime

WORKDIR /app/demo
COPY --from=builder /app/mtc-cli        /app/mtc-cli
COPY --from=builder /app/demo/website-server /app/demo/website-server
COPY --from=builder /app/demo/setup.sh  /app/demo/setup.sh
COPY --from=builder /app/demo/static    /app/demo/static

RUN chmod +x /app/mtc-cli /app/demo/website-server /app/demo/setup.sh

EXPOSE 8443

# Entrypoint for demo: run setup then start server
CMD ["/bin/bash", "-c", "./setup.sh && ./website-server"]

# ==========================================
# Playground Runtime Stage
# ==========================================
FROM runtime-base AS playground-runtime

WORKDIR /app/playground
COPY --from=builder /app/playground/playground-server /app/playground/playground-server
COPY --from=builder /app/playground/static            /app/playground/static

# ── DigiCert playground standalone tools
COPY --from=builder /tmp/demo-embedded-cert /usr/local/bin/demo-embedded-cert
COPY --from=builder /tmp/mtc-verify-cert    /usr/local/bin/mtc-verify-cert
COPY --from=builder /tmp/mtc-conformance    /usr/local/bin/mtc-conformance
COPY --from=builder /tmp/mtc-interop        /usr/local/bin/mtc-interop

RUN chmod +x /app/playground/playground-server \
             /usr/local/bin/demo-embedded-cert \
             /usr/local/bin/mtc-verify-cert \
             /usr/local/bin/mtc-conformance \
             /usr/local/bin/mtc-interop

EXPOSE 8444

ENV PLAYGROUND_BIN_DIR=/usr/local/bin
CMD ["./playground-server"]

