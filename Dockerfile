####################################################################################################
# base
####################################################################################################
FROM alpine:3.12.3 as base
ARG ARCH=amd64

RUN apk update && apk upgrade && \
    apk add ca-certificates && \
    apk --no-cache add tzdata

COPY dist/apache-pulsar-sink-go-${ARCH} /bin/apache-pulsar-sink-go
RUN chmod +x /bin/apache-pulsar-sink-go

####################################################################################################
# apache-pulsar-sink-go
####################################################################################################
FROM scratch as apache-pulsar-sink-go
ARG ARCH
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base /bin/apache-pulsar-sink-go /bin/apache-pulsar-sink-go
ENTRYPOINT [ "/bin/apache-pulsar-sink-go" ]

####################################################################################################
# testbase
####################################################################################################
FROM alpine:3.17 as testbase
RUN apk update && apk upgrade && \
    apk add ca-certificates && \
    apk --no-cache add tzdata

COPY dist/e2eapi /bin/e2eapi
RUN chmod +x /bin/e2eapi


####################################################################################################
# testapi
####################################################################################################
FROM scratch AS e2eapi
COPY --from=testbase /bin/e2eapi .
ENTRYPOINT ["/e2eapi"]
