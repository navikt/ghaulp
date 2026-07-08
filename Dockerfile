FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /ghaulp .

FROM cgr.dev/chainguard/static:latest

COPY --from=builder /ghaulp /ghaulp

ENTRYPOINT ["/ghaulp"]
