# syntax=docker/dockerfile:1

# --- builder ---------------------------------------------------------------
FROM golang:1.27-bookworm AS builder
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
WORKDIR /app

# Camada de dependências isolada: só rebuilda se os manifests mudarem.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0: lib/pq e go-sql-driver/mysql são Go puro, então o
# binário sai estático e roda em scratch.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -trimpath \
      -ldflags "-s -w \
        -X github.com/Jeffersonmf/aurora-exporter/internal/version.Version=${VERSION} \
        -X github.com/Jeffersonmf/aurora-exporter/internal/version.Commit=${COMMIT} \
        -X github.com/Jeffersonmf/aurora-exporter/internal/version.Date=${DATE}" \
      -o /out/aurora-exporter ./cmd/aurora-exporter

# --- runtime -----------------------------------------------------------
# scratch + certificados de CA (sslmode=verify-full para Redshift/RDS)
# + usuário não-root. Sem shell, sem libc.
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/aurora-exporter /aurora-exporter
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/aurora-exporter"]
CMD ["serve", "-config", "/etc/aurora-exporter/config.yaml"]
