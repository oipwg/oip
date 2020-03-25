#!/usr/bin/env bash

echo "Building oip proto files"
cd $GOPATH/src/github.com/oipwg/oip/proto/oip
protoc --go_out=$GOPATH/src multipart.proto pubkey.proto signedMessage.proto txid.proto

echo "Building oip5 proto files"
cd $GOPATH/src/github.com/oipwg/oip/proto/oip5
protoc --go_out=$GOPATH/src -I=. -I=$GOPATH/src/github.com/oipwg/oip/proto/oip \
  edit.proto NormalizeRecord.proto NormalizeRecord.proto oip5.proto Record.proto RecordTemplateProto.proto

echo "Building oip5 template proto files"
cd $GOPATH/src/github.com/oipwg/oip/proto/oip5/templates
protoc --go_out=$GOPATH/src tmpl_433C2783.proto

echo "Building historian proto files"
cd $GOPATH/src/github.com/oipwg/oip/proto/historian
protoc --go_out=$GOPATH/src historian.proto
