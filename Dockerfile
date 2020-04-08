FROM golang as builder

WORKDIR /build
COPY . .

RUN make build

FROM alpine

RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates
RUN apk update && \
    apk upgrade && \
    apk add bash && \
    apk add curl

COPY --from=builder /build/clickhouse-backup/clickhouse-backup /bin/clickhouse-backup
RUN chmod +x /bin/clickhouse-backup

ENTRYPOINT [ "/bin/clickhouse-backup" ]
CMD [ "--help" ]
