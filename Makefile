.PHONY: gen-user migrate-user

MIGRATIONS_PATH = migrations/user

ifeq ($(OS),Windows_NT)
    MKDIR_CMD = if not exist gen\go\user mkdir gen\go\user
    MKDIR_MIGRATIONS = if not exist $(MIGRATIONS_PATH) mkdir $(MIGRATIONS_PATH)
else
    MKDIR_CMD = mkdir -p gen/go/user
    MKDIR_MIGRATIONS = mkdir -p $(MIGRATIONS_PATH)
endif

# make gen-user
gen-user:
	$(MKDIR_CMD)
	protoc --proto_path=protos/user --go_out=gen/go/user --go_opt=paths=source_relative --go-grpc_out=gen/go/user --go-grpc_opt=paths=source_relative user.proto

# make migrate-user name=create_users_table
migrate-user:
	$(MKDIR_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

run-compose:
	docker-compose up -d --build

stop-compose:
	docker-compose down