#!/bin/sh

# Generate SP metadata
mdpath=./tmp/saml-test-${USERø}-${HOST}.xml
curl localhost:8090/saml/metadata > $mdpath
# push sp metadata
curl -i -F userfile=@$mdpath https://www.testshib.org/procupload.php