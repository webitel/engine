protoc -I/usr/local/include -I./protos \
  --go-grpc_out=./api --go-grpc_opt=paths=source_relative \
  --go_opt=paths=source_relative --go_out=./api \
  ./protos/*.proto