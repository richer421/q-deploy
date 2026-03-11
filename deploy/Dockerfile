# ---------- 构建阶段 ----------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache \
    gcc \
    musl-dev \
    librdkafka-dev \
    pkgconf

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -tags musl -o /app/{{ .ProjectName }} .

# ---------- 运行阶段 ----------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata librdkafka

WORKDIR /app
COPY --from=builder /app/{{ .ProjectName }} .
COPY conf/config.yaml ./conf/config.yaml

EXPOSE 8080

ENTRYPOINT ["./{{ .ProjectName }}"]
CMD ["server"]
