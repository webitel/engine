protoc -I/usr/local/include -I./protos -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis  \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
   --go_out=plugins=grpc:./api ./protos/*.proto
