FROM golang:latest as builder
LABEL maintainer "ohh <ohhfrancois@free.fr>"

RUN mkdir /build
ADD *.go /build/
RUN rm -f /build/*_test.go

WORKDIR /build

RUN set -x              && \
    go get -d -v .      && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o SAMLAuthnRequester 

# Docker run Golang app
#FROM alpine:latest
FROM scratch
LABEL maintainer "ohh <ohhfrancois@free.fr>"

ENV SAMLRQT_IDP_URLMETADATA Foo
ENV SAMLRQT_SP_ROOTURL Foo
ENV SAMLRQT_SP_ID Foo
ENV SAMLRQT_CERT_FILE Foo
ENV SAMLRQT_CERT_KEY Foo

VOLUME /app/certificates

WORKDIR /app

EXPOSE 8090

COPY --from=builder /build/SAMLAuthnRequester /app/SAMLAuthnRequester

# executable
ENTRYPOINT ["/app/SAMLAuthnRequester"]
# arguments that can be overridden
# No parameter yet
# CMD [ "8090r" ]