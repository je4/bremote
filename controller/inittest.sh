#!/bin/bash

# client01

curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"title":"Client 01","wsAddr":"wss://localhost:443/ws"}' \
  --insecure \
  https://localhost:8443/kvstore/client01/wstest.html
echo ""


# client02

curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"title":"Client 02","wsAddr":"wss://localhost:444/ws"}' \
  --insecure \
  https://localhost:8443/kvstore/client02/wstest.html
echo ""


# groups

curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"instance":"client01"}' \
  --insecure \
  https://localhost:8443/groups/test
echo ""
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"instance":"client02"}' \
  --insecure \
  https://localhost:8443/groups/test
echo ""
