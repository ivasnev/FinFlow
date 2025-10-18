generate:
	oapi-codegen  -package api -generate types -o pkg/api/types.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate client -o pkg/api/client.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate gin-server -o pkg/api/server.gen.go pkg/api/openapi.yaml
	oapi-codegen  -package api -generate spec -o pkg/api/spec.gen.go pkg/api/openapi.yaml
