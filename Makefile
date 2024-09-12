swagger-install:
	which swagger || go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger:
	swagger generate spec -o ./swagger.yaml --scan-models
