#!/usr/bin/env bash

URL=http://localhost:1337/v1/storage/files
BUCKET=default

FILE_ID=55af1e60-0f28-454e-885e-ea6aab2bb288
ETAG=\"588be441fe7a59460850b0aa3e1c5a65\"

# we sleep for 1s to make sure a drift in the clocks between client/server doesn't
# lead to a JWTIssuedAtFuture error
sleep 1

curl $URL/ \
  -v \
  -H "Content-Type: multipart/form-data" \
  -F "bucket-id=$BUCKET" \
  -F "file[]=@client/testdata/alphabet.txt" \
  -F "file[]=@hasura-storage.yaml" \
  -F "file[]=@README.md" \
  -F "file[]=@image/testdata/nhost.jpg"
