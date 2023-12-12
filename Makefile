run-db:
	docker run \
    --rm --name postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=password \
    -e POSTGRES_DB=postgres \
    -p 5432:5432 \
    -d postgres:latest

proto:
	protoc ./proto/*.proto --go-grpc_out=pkg --go_out=pkg