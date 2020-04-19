# -----------------------------------------------------------------------------
# Build image
# -----------------------------------------------------------------------------
FROM golang:alpine AS backend

RUN apk update && \
    apk add --no-cache git gcc musl-dev

WORKDIR /go/src/github.com/itglobal/slack-uptime-monitor-cache
COPY go.mod .
COPY go.sum .
RUN go mod download

WORKDIR /go/src/github.com/itglobal/slack-uptime-monitor

COPY . .
RUN go get
RUN go build -o /out/slack-uptime-monitor

# -----------------------------------------------------------------------------
# Runtime image
# -----------------------------------------------------------------------------
FROM alpine:latest
RUN apk update && apk add --no-cache curl
WORKDIR /app
VOLUME /app/var
ENV VAR_DIR=/app/var
ENV LISTEN_ADDRESS=0.0.0.0:5000
COPY --from=backend /out/slack-uptime-monitor /app/slack-uptime-monitor
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD curl -q http://localhost:5000
ENTRYPOINT [ "/app/slack-uptime-monitor" ]
