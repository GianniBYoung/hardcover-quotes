FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
RUN CGO_ENABLED=0 go build -o /app/hcq ./cmd/hcq

FROM cgr.dev/chainguard/static:latest
COPY --from=builder /app/hcq /usr/local/bin/hcq
USER nonroot
ENTRYPOINT ["/usr/local/bin/hcq"]

