.PHONY: all php go clean

#PROTOC=/usr/local/src/grpc/bins/opt/protobuf/protoc
PROTOC=/usr/local/bin/protoc
GRPC_PHP_PLUGIN=/usr/local/bin/grpc_php_plugin
#GRPC_PHP_PLUGIN=$(PROTOC)
GRPC_GO_PLUGIN=/mnt/hgfs/go/bin/protoc-gen-go

GRPC_GO_OUT=.
GRPC_PHP_OUT=.
GRPC_JAVA_OUT=.


PROTO_PATH=-I . -I /usr/local/src/grpc/third_party/protobuf/src

#PROTOC_PHP_PARAM=--plugin=protoc-gen-grpc=$(GRPC_PHP_PLUGIN) --php_out=mediaserver-php --grpc_out=mediaserver-php
PROTOC_PHP_PARAM=--php_out=mediaserver-php --grpc_out=$(GRPC_PHP_OUT) --plugin=protoc-gen-grpc=$(GRPC_PHP_PLUGIN)
PROTOC_GO_PARAM=--go_out=plugins=grpc:$(GRPC_GO_OUT)
PROTOC_JAVA_PARAM_SERVICE=--plugin=protoc-gen-grpc-java=/usr/local/bin/protoc-gen-grpc-java --grpc-java_out=lite:$(GRPC_JAVA_OUT)
PROTOC_JAVA_PARAM_OBJ=--java_out=$(GRPC_JAVA_OUT)

all: types.pb.go client.pb.go proxy.pb.go controller.pb.go http.pb.go

%.pb.go: %.proto
	$(PROTOC) $(PROTO_PATH) $(PROTOC_GO_PARAM) $<
