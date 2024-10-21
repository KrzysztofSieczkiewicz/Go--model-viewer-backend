swagger-install:
	which swagger || go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger:
	@echo Generating Swagger spec for FilesService...
	swagger generate spec -o ./swagger.yaml --scan-models