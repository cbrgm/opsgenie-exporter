FROM alpine:latest
RUN apk add --update ca-certificates

ADD opsgenie-exporter /usr/bin/opsgenie_exporter

EXPOSE 9212

ENTRYPOINT ["/usr/bin/opsgenie_exporter"]
