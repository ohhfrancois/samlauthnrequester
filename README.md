SAML Authn Requester
------------------------------

SAML Requester is a small program GOLANG to send SAML AuthnRequest to an IDP 

# Table of Contents

- [Table of Contents](#table-of-contents)
- [How to Use standalone](#how-to-use-standalone)
  - [Generate your certificate to sign the request](#generate-your-certificate-to-sign-the-request)
  - [Initialize environment variables](#initialize-environment-variables)
  - [Launch SAMLRequester executable](#launch-samlrequester-executable)
  - [Example](#example)
- [How to Use docker](#how-to-use-docker)
  - [Standalone Container](#standalone-container)
  - [Compose Container](#compose-container)
- [Development](#development)
- [Testing](#testing)
- [Release](#release)
  - [Docker Release](#docker-release)
- [Tips](#tips)
  - [Generate your spProvide Metadata to testshib.org](#generate-your-spprovide-metadata-to-testshiborg)
  - [Generate the pfx cert file for windows from your cert and keyfile](#generate-the-pfx-cert-file-for-windows-from-your-cert-and-keyfile)

# How to Use standalone

## Generate your certificate to sign the request

Each service provider must have an self-signed X.509 key pair established. You can generate your own with something like this:

```bash
openssl req -x509 -newkey rsa:2048 -keyout myservice.key -out myservice.cert -days 365 -nodes -subj "/CN=myservice.example.com"
```

## Initialize environment variables

- SAMLRQT_IDP_URLMETADATA : IDP URL Metadata file

- SAMLRQT_SP_ROOTURL : SP Root URL
- SAMLRQT_SP_ID: Service Provider ID

- SAMLRQT_CERT_FILE: Path to the sign cert file
- SAMLRQT_CERT_KEY: Path to the sign key file

## Launch SAMLRequester executable

```bash
$GOLANG/bin/SAMLAuthnRequester
```

## Example

```bash
export SAMLRQT_IDP_URLMETADATA=https://samltest.id/saml/idp

export SAMLRQT_SP_ID=myapp
export SAMLRQT_SP_ROOTURL=http://localhost:8090/myapp

export SAMLRQT_CERT_FILE=Foo-CERT-FILE.cert
export SAMLRQT_CERT_KEY=Foo-CERT-KEY.key

$GOLANG/bin/SAMLAuthnRequester
```

# How to Use docker

## Standalone Container

The docker registry is on AWS, need to docker login before pull

```bash
docker run -v $PWD/certificates:/app/certificates --env-file ./docker.envfile --rm 429815655062.dkr.ecr.us-west-2.amazonaws.com/psg-france/samlauthnrequester:latest
```

## Compose Container

generate you docker-compose.yml file

```yaml
version: '3.3'

services:
  SAMLAuthnRequester:
    image: 429815655062.dkr.ecr.us-west-2.amazonaws.com/psg-france/samlauthnrequester:latest
    container_name: SAMLAuthnRequester
    labels:
      traefik.enable: true
      traefik.http.routers.SAMLAuthnRequester.rule: Host(`mydomain.example.com` && PathPrefix(`/saml-requester`))
      traefik.http.routers.SAMLAuthnRequester.entrypoints: web, websecure
    expose:
      - 8090
    environment:   
      - HTTPS_PROXY=http://myProxy:myproxyport
      - https_proxy=http://myProxy:myproxyport
      - SAMLRQT_IDP_URLMETADATA=IDP-URL-Metadata
      - SAMLRQT_SP_ROOTURL=SP-Root-URL
      - SAMLRQT_SP_ID=SP-ID
      - SAMLRQT_CERT_FILE=/certificates/mycert.cert
      - SAMLRQT_CERT_KEY=/certificates/mycert.key
    volumes:
      - $PWD/certificates:/certificates
    networks:
      - traefik

networks:
  traefik:
    external: true
```

Create traefik Network if needed and launch the container

```bash
docker network create traefik
docker-compose up -d SAMLAuthnRequester
```

# Development

```bash
go init
make vendor
```

develop ...

```bash
make run
```

# Testing

Will :
- build local container
- generate certificate
- generate sp metadata
- add saml sp metadata to test IDp
- launch container
- request saml-requester
- redirect to http://localhost:8090/saml


```bash
make validate
```


# Release

## Docker Release

Modify the release version in file docker-manifest.json
```bash
make build-ecr
```

# Tips

## Generate your spProvide Metadata to testshib.org

```bash
mdpath=saml-test-$USER-$HOST.xml
curl localhost:8090/saml/metadata > $mdpath
curl -i -F userfile=@$mdpath https://www.testshib.org/procupload.php
```

## Generate the pfx cert file for windows from your cert and keyfile

```bash
openssl pkcs12 -export -out certificates/myservice.pfx -inkey certificates/myservice.key -in certificates/myservice.crt 
```
