.PHONY: gen-user migrate-user gen-flight migrate-flight gen-payment migrate-payment

MIGRATIONS_USER_PATH = migrations/user
MIGRATIONS_FLIGHT_PATH = migrations/flight
MIGRATIONS_PAYMENT_PATH = migrations/payment

ifeq ($(OS),Windows_NT)
    # User Service
    MKDIR_USER_GEN = if not exist gen\go\user mkdir gen\go\user
    MKDIR_USER_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_USER_PATH)) mkdir $(subst /,\,$(MIGRATIONS_USER_PATH))

    # Flight Service
    MKDIR_FLIGHT_GEN = if not exist gen\go\flight mkdir gen\go\flight
    MKDIR_FLIGHT_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_FLIGHT_PATH)) mkdir $(subst /,\,$(MIGRATIONS_FLIGHT_PATH))

	# Payment Service
    MKDIR_PAYMENT_GEN = if not exist gen\go\payment mkdir gen\go\payment
    MKDIR_PAYMENT_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_PAYMENT_PATH)) mkdir $(subst /,\,$(MIGRATIONS_PAYMENT_PATH))
else
    # User Service
    MKDIR_USER_GEN = mkdir -p gen/go/user
    MKDIR_USER_MIGRATIONS = mkdir -p $(MIGRATIONS_USER_PATH)

    # Flight Service
    MKDIR_FLIGHT_GEN = mkdir -p gen/go/flight
    MKDIR_FLIGHT_MIGRATIONS = mkdir -p $(MIGRATIONS_FLIGHT_PATH)

	# Payment Service
	MKDIR_PAYMENT_GEN = mkdir -p gen/go/payment
    MKDIR_PAYMENT_MIGRATIONS = mkdir -p $(MIGRATIONS_PAYMENT_PATH)
endif

run-compose:
	docker-compose up -d --build

stop-compose:
	docker-compose down

# make gen-user
gen-user:
	$(MKDIR_USER_GEN)
	protoc --proto_path=protos/user --go_out=gen/go/user --go_opt=paths=source_relative --go-grpc_out=gen/go/user --go-grpc_opt=paths=source_relative user.proto

# make migrate-user name=create_users_table
migrate-user:
	$(MKDIR_USER_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

# make gen-flight
gen-flight:
	$(MKDIR_FLIGHT_GEN)
	protoc --proto_path=protos/flight --go_out=gen/go/flight --go_opt=paths=source_relative --go-grpc_out=gen/go/flight --go-grpc_opt=paths=source_relative flight.proto

# make migrate-flight name=create_flights_table
migrate-flight:
	$(MKDIR_FLIGHT_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_FLIGHT_PATH) -seq $(name)

# make gen-payment
gen-payment:
	$(MKDIR_PAYMENT_GEN)
	protoc --proto_path=protos/payment --go_out=gen/go/payment --go_opt=paths=source_relative --go-grpc_out=gen/go/payment --go-grpc_opt=paths=source_relative payment.proto

# make migrate-payment name=init_schema
migrate-payment:
	$(MKDIR_PAYMENT_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_PAYMENT_PATH) -seq $(name)