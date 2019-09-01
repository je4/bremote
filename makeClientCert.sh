#!/bin/bash
# call this script with an email address (valid or not).
# like:
# ./makecert.sh foo@foo.com

if [ "$1" == "" ]; then
    echo "Need email as first argument"
    exit 1
fi

if [ "$2" == "" ]; then
    echo "Need prefix as second argument"
    exit 1
fi


EMAIL=$1
PREFIX=$2

rm -rf certs
mkdir certs
cd certs

echo "make CA"
PRIVKEY="digma38975524"
openssl req -new -x509 -days 365 -keyout ${PREFIX}ca.key -out ${PREFIX}ca.pem -subj "/C=CH/ST=Basel/L=Earth/O=FHNW/OU=HGK/OU=DIGMA/CN=www.fhnw.ch/emailAddress=ict.help.hgk@fhnw.ch" -passout pass:$PRIVKEY

echo "make server cert"
openssl req -new -nodes -x509 -out ${PREFIX}server.pem -keyout ${PREFIX}server.key -days 3650 -subj "/C=CH/ST=Basel/L=Earth/O=FHNW/OU=HGK/OU=DIGMA/CN=www.fhnw.ch/emailAddress=${EMAIL}"

echo "make client cert"
#openssl req -new -nodes -x509 -out ${PREFIX}client.pem -keyout ${PREFIX}client.key -days 3650 -subj "/C=CH/ST=Basel/L=Earth/O=FHNW/OU=HGK/OU=DIGMA/CN=www.fhnw.ch/emailAddress=${EMAIL}"

openssl genrsa -out ${PREFIX}client.key 2048
echo "00" > ${PREFIX}ca.srl
openssl req -sha1 -key ${PREFIX}client.key -new -out ${PREFIX}client.req -subj "/C=CH/ST=Basel/L=Earth/O=FHNW/OU=HGK/OU=DIGMA/OU=fhnw.ch/CN=bremote/emailAddress=${EMAIL}"
# Adding -addtrust clientAuth makes certificates Go can't read
openssl x509 -req -days 365 -in ${PREFIX}client.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${PREFIX}client.pem # -addtrust clientAuth

openssl x509 -extfile ../openssl.conf -extensions ssl_client -req -days 365 -in ${PREFIX}client.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${PREFIX}client.pem