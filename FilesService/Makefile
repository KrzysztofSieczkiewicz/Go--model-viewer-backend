swagger-install:
	which swagger || go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger:
	@echo Generating Swagger spec for FilesService...
	swagger generate spec -o ./swagger.yaml --scan-models

protos:
	protoc -I protos/ --go_out=. --go-grpc_out=. protos/files.proto
	protoc -I protos/ protos/files.proto --go_out=plugins=grpc:protos/files