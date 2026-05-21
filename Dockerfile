FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/kube-dash ./cmd/kube-dash

FROM ghcr.io/kongken/go-runtime:main

COPY --from=builder /out/kube-dash /app/bin/kube-dash

ENTRYPOINT ["/app/bin/kube-dash"]
