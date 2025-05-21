test:
	go test -v -cover -short ./...

# mock:
# 	mockgen -source="./distributor.go" -package=mockwk -destination="mock/distributor_mock.go"
# 	mockgen  . Store

db_docs:
	docker run --rm -v $(current_dir)/doc:/app/data -w /app/data node_docs dbdocs build ./db.dbml

db_schema:
	docker run --rm -v $(current_dir)/doc:/app/data -w /app/data node_docs dbml2sql --postgres -o schduler_schema.sql db.dbml

protoc:
	powershell -Command "Remove-Item -Path 'api/pb/*.go' -Force"
	protoc   --grpc-gateway_out api/pb \
	--proto_path=C:/Users/royce/go/pkg/mod/github.com/protocolbuffers/protobuf@v4.24.4+incompatible/src \
	--proto_path=proto  --go_out=api/pb  --go_opt=paths=source_relative  --grpc-gateway_opt=paths=source_relative \
	--go-grpc_out=api/pb --go-grpc_opt=paths=source_relative \
	proto/*.proto

evans:
	docker run -it --rm -v $(current_dir):/mount:ro ghcr.io/ktr0731/evans:latest --host host.docker.internal --port 9090 -r repl

run:
	go run cmd/server/main.go

.PHONY: test db_docs db_schema protoc evans run
 