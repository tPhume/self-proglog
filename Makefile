compile:
	protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf \
	--gogo_out=plugins=grpc:. ./api/v1/log.proto