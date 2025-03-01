FROM alpine:latest

LABEL maintainer="Artem Klevtsov <a.a.klevtsov@gmail.com>"

COPY redpanda-connect /usr/bin/redpanda-connect

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown appuser:appgroup /app/

USER appuser

EXPOSE 4195

ENTRYPOINT ["/usr/bin/redpanda-connect"]

CMD ["run"]
