FROM cgr.dev/chainguard/go:latest
COPY cmd/* /app
COPY go.mod /app
COPY go.sum /app
WORKDIR /app
RUN go build -o hcq /app
RUN install -Dm711 /app/hcq /usr/local/bin/hcq
ENTRYPOINT hcq
