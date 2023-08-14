#!/bin/sh
#set -x

dst=api
#dst=api/contacts
src=proto/contacts

# ensure target dir exists
mkdir -p $dst

#protoc -I $src -I proto \
protoc -I proto \
  --go_opt=paths=source_relative --go_out=$dst \
  --micro_out=plugins=grpc,paths=source_relative:$dst \
  $src/*.proto