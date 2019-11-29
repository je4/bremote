#!/bin/bash
# call this script with an email address (valid or not).
# like:
# ./makecert.sh foo@foo.com

if [ "$1" == "" ]; then
    echo "Need email as first argument"
    exit 1
fi

if [ "$2" == "" ]; then
    echo "Need file prefix as second argument"
    exit 1
fi


EMAIL=$1
PREFIX=$2

BASEDN="/C=CH/ST=Basel/O=info-age/OU=display01"

mkcert() {
  machine=$1
  type=$2

  certfile="${PREFIX}${machine}"
  echo ${certfile} ++++
  openssl genrsa -out ${certfile}.key 2048
  openssl req -sha1 -key ${certfile}.key -new -out ${certfile}.req -subj "/L=${type}${BASEDN}/CN=${machine}/emailAddress=${EMAIL}"
  # Adding -addtrust clientAuth makes certificates Go can't read
  #openssl x509 -req -extfile <(printf "subjectAltName=DNS:${type}") -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem # -addtrust clientAuth
  # uuid: 7b721292-cc3d-4855-99e1-262444bce988 == type
  openssl x509 -req -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem


  openssl x509 -extfile ../openssl.conf -extensions ssl_client -req -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem
}

#rm -f certs/${PREFIX}*
#mkdir certs
cd certs

echo "00" > ${PREFIX}ca.srl

echo "make CA"
PRIVKEY="info-age21654968473214dsD"
openssl req -new -x509 -days 365 -keyout ${PREFIX}ca.key -out ${PREFIX}ca.pem -subj "${BASEDN}/CN=ca/emailAddress=juergen@info-age.net" -passout pass:$PRIVKEY

echo "make localhost cert for https access"
openssl req -new -nodes -x509 -out ${PREFIX}localhost.pem -keyout ${PREFIX}localhost.key -days 3650 -subj "${BASEDN}/CN=localhost/emailAddress=${EMAIL}"

#echo "make client cert"
#openssl req -new -nodes -x509 -out ${PREFIX}client.pem -keyout ${PREFIX}client.key -days 3650 -subj "/C=CH/ST=Basel/L=Earth/O=FHNW/OU=HGK/OU=DIGMA/CN=www.fhnw.ch/emailAddress=${EMAIL}"

