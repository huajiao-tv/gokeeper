target: gen

_cleango:
	@rm -rf go/*

_gengo:
	@protoc -Iproto \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
		-I${GOPATH}/src/github.com/lnnujxxy/protoc-gen-validate \
		--go_out=plugins=grpc:go \
		--grpc-gateway_out=logtostderr=true:go \
		--validate_out="lang=go:go" \
		proto/*.proto

gen: clean _gengo

clean: _cleango