FROM golang:alpine AS build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

ENV GOCACHE=/root/.cache/go-build

RUN apk add go-task-task git

RUN --mount=type=cache,target="/root/.cache/go-build" \
    task go:build

FROM alpine:latest AS package

LABEL maintainer="Artem Klevtsov <a.klevtsov@gmail.com>"

COPY --from=build /app/redpanda-connect-yandex-metrika /usr/bin/redpanda-connect-yandex-metrika

WORKDIR /app

RUN addgroup appuser && adduser -D -G appuser appuser
RUN chown appuser:appuser /app/

USER appuser

EXPOSE 4195

ENTRYPOINT ["/usr/bin/redpanda-connect-yandex-metrika"]

CMD ["run"]
