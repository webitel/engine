protoc -I/usr/local/include -I./protos -I$GRPC_GATEWAY/third_party/googleapis  \
  -I$GRPC_GATEWAY \
   --go_out=plugins=grpc:./api ./protos/*.proto
