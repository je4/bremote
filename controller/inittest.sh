#!/bin/bash

# client01

echo set https://localhost:8443/kvstore/client01/wstest.html to '{"title":"Client 01","wsAddr":"wss://localhost:443/ws"}'
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"title":"Client 01","wsAddr":"wss://localhost:443/ws"}' \
  --insecure \
  https://localhost:8443/kvstore/client01/wstest.html
echo ""


# client02

echo set https://localhost:8443/kvstore/client02/wstest.html to '{"title":"Client 02","wsAddr":"wss://localhost:444/ws"}'
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"title":"Client 02","wsAddr":"wss://localhost:444/ws"}' \
  --insecure \
  https://localhost:8443/kvstore/client02/wstest.html
echo ""


# mac02

echo set https://localhost:8443/kvstore/mac02/wstest.html to '{"title":"Mac 02","wsAddr":"wss://localhost:443/ws"}'
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"title":"Mac 02","wsAddr":"wss://localhost:443/ws"}' \
  --insecure \
  https://localhost:8443/kvstore/mac02/wstest.html
echo ""


# groups

echo add client01 to group test
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"instance":"client01"}' \
  --insecure \
  https://localhost:8443/groups/test
echo ""
echo add client02 to group test
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"instance":"client02"}' \
  --insecure \
  https://localhost:8443/groups/test
echo ""
echo add mac02 to group test
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"instance":"mac02"}' \
  --insecure \
  https://localhost:8443/groups/test
echo ""

# load mac02 website
echo navigate mac02 to https://localhost:443/controller01/templates/wstest.html
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"url": "https://localhost:443/controller01/templates/wstest.html","nextstatus":"wstest"}' \
  --insecure \
  https://localhost:8443/client/mac02/navigate
echo ""
