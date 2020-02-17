FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN go build -o coyote .

FROM alpine
RUN adduser -S -D -H -h /app appuser
CMD chown -R $(stat -c "%u:%g" /tmp) /shared
USER appuser
COPY --chown=appuser:nogroup --from=builder /build/coyote /usr/local/bin/

COPY --chown=appuser:nogroup coyote.yml coyote-skip.yml /app/files/

WORKDIR /app
CMD ["coyote -c ./files/*"]
SHELL ["/bin/bash", "-c"]
