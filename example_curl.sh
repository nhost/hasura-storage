#!/usr/bin/env bash

URL=http://localhost:8000/v1/storage
AUTH="Authorization: Bearer $(make dev-jwt)"
BUCKET=default

FILE_ID=55af1e60-0f28-454e-885e-ea6aab2bb288
ETAG=\"588be441fe7a59460850b0aa3e1c5a65\"

output=`curl $URL/files/ \
  -v \
  -H "Content-Type: multipart/form-data" \
  -H "$AUTH" \
  -F "bucket-id=$BUCKET" \
  -F "metadata[]={};type=application/json" \
  -F "file[]=@client/testdata/alphabet.txt" \
  -F "metadata[]={\"id\":\"7982873d-8e89-4321-ab86-00f80a168c5a\", \"name\":\"a_folder/config.yaml\"};type=application/json" \
  -F "file[]=@hasura-storage.yaml" \
  -F "metadata[]={\"id\":\"faa80d51-07c7-4268-942d-8f092c98c71a\", \"name\":\"a_folder/docs.md\"};type=application/json" \
  -F "file[]=@README.md" \
  -F "metadata[]={\"id\":\"$FILE_ID\", \"name\":\"b_folder/logo.jpg\"};type=application/json" \
  -F "file[]=@image/testdata/nhost.jpg"`

echo $output | jq

time curl -v -o nhost.jpg $URL/files/${FILE_ID} \
      -H "$AUTH"

time curl -v -o nhost.jpg $URL/files/${FILE_ID}?x=600\&y\=200\&q=50\&b=5 \
      -H "$AUTH"


curl -v $URL/buckets/ \
  -H "$AUTH" | jq


curl -v $URL/buckets/default \
  -H "$AUTH" | jq


curl -v $URL/buckets/default/list-files \
  -H "$AUTH" | jq


curl -v $URL/buckets/default/list-files?filter_files_regex="^a_folder/.*" \
  -H "$AUTH" | jq
