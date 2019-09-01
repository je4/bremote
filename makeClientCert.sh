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

mkcert() {
  machine = $1
  type = $2

  certfile="${PREFIX}${machine}"
  echo ${certfile}
  openssl genrsa -out ${certfile}.key 2048
  openssl req -sha1 -key ${certfile}.key -new -out ${certfile}.req -subj "/C=CH/ST=Basel/O=info-age/OU=display01/CN=${machine}/emailAddress=${EMAIL}" -addext "type = ${2}"
  # Adding -addtrust clientAuth makes certificates Go can't read
  openssl x509 -req -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem # -addtrust clientAuth

  openssl x509 -extfile ../openssl.conf -extensions ssl_client -req -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem
}

rm -f certs/${PREFIX}*
#mkdir certs
cd certs

echo "make CA"
PRIVKEY="info-age21654968473214dsD"
openssl req -new -x509 -days 365 -keyout ${PREFIX}ca.key -out ${PREFIX}ca.pem -subj "/C=CH/ST=Basel/O=info-age/OU=display01/CN=ca/emailAddress=juergen@info-age.net" -passout pass:$PRIVKEY

#echo "make server cert"
#openssl req -new -nodes -x509 -out ${PREFIX}server.pem -keyout ${PREFIX}server.key -days 3650 -subj "/C=CH/ST=Basel/O=info-age/OU=display01/CN=proxy/emailAddress=${EMAIL}"

#echo "make client cert"
#openssl req -new -nodes -x509 -out ${PREFIX}client.pem -keyout ${PREFIX}client.key -days 3650 -subj "/C=CH/ST=Basel/L=Earth/O=FHNW/OU=HGK/OU=DIGMA/CN=www.fhnw.ch/emailAddress=${EMAIL}"

declare -a machines=()

machines+=("master")
machines+=("proxy")
#machines+=("proxy02")
machines+=("controller01")
machines+=("controller02")

for i in {1..99}
do
  num=$(printf "%03d" $i)
  machine=ba14nc21${num}
  machines+=($machine)
done

echo "00" > ${PREFIX}ca.srl

for machine in "${machines[@]}"
do
  certfile="${PREFIX}${machine}"
  echo ${certfile}
  openssl genrsa -out ${certfile}.key 2048
  openssl req -sha1 -key ${certfile}.key -new -out ${certfile}.req -subj "/C=CH/ST=Basel/O=info-age/OU=display01/CN=${machine}/emailAddress=${EMAIL}"
  # Adding -addtrust clientAuth makes certificates Go can't read
  openssl x509 -req -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem # -addtrust clientAuth

  openssl x509 -extfile ../openssl.conf -extensions ssl_client -req -days 365 -in ${certfile}.req -CA ${PREFIX}ca.pem -CAkey ${PREFIX}ca.key -passin pass:$PRIVKEY -out ${certfile}.pem

done