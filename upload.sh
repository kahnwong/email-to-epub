#!/bin/bash

source .env

json_object='{
    "path": "/Apps/Dropbox PocketBook/output.epub",
    "mode": "add",
    "autorename": true,
    "mute": false
}'

json_string=$(echo "$json_object" | jq . | jq -r tostring)
echo "$json_string"

curl -X POST https://content.dropboxapi.com/2/files/upload \
	--header "Authorization: Bearer ${DROPBOX_TOKEN}" \
	--header "Content-Type: application/octet-stream" \
	--header "Dropbox-API-Arg: ${json_string}" \
	--data-binary @"$(pwd)/output.epub"
