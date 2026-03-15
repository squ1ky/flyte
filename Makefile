.PHONY: gen-user migrate-user gen-flight migrate-flight gen-payment migrate-payment gen-booking migrate-booking run-compose start-compose

MIGRATIONS_USER_PATH = migrations/user
MIGRATIONS_FLIGHT_PATH = migrations/flight
MIGRATIONS_PAYMENT_PATH = migrations/payment
MIGRATIONS_BOOKING_PATH = migrations/booking

ifeq ($(OS),Windows_NT)
    # User Service
    MKDIR_USER_GEN = if not exist gen\proto\user mkdir gen\proto\user
    MKDIR_USER_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_USER_PATH)) mkdir $(subst /,\,$(MIGRATIONS_USER_PATH))

    # Flight Service
    MKDIR_FLIGHT_GEN = if not exist gen\proto\flight mkdir gen\proto\flight
    MKDIR_FLIGHT_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_FLIGHT_PATH)) mkdir $(subst /,\,$(MIGRATIONS_FLIGHT_PATH))

	# Payment Service
    MKDIR_PAYMENT_GEN = if not exist gen\proto\payment mkdir gen\proto\payment
    MKDIR_PAYMENT_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_PAYMENT_PATH)) mkdir $(subst /,\,$(MIGRATIONS_PAYMENT_PATH))

    # Booking Service
    MKDIR_BOOKING_GEN = if not exist gen\proto\booking mkdir gen\proto\booking
    MKDIR_BOOKING_MIGRATIONS = if not exist $(subst /,\,$(MIGRATIONS_BOOKING_PATH)) mkdir $(subst /,\,$(MIGRATIONS_BOOKING_PATH))
else
    # User Service
    MKDIR_USER_GEN = mkdir -p gen/proto/user
    MKDIR_USER_MIGRATIONS = mkdir -p $(MIGRATIONS_USER_PATH)

    # Flight Service
    MKDIR_FLIGHT_GEN = mkdir -p gen/proto/flight
    MKDIR_FLIGHT_MIGRATIONS = mkdir -p $(MIGRATIONS_FLIGHT_PATH)

	# Payment Service
	MKDIR_PAYMENT_GEN = mkdir -p gen/proto/payment
    MKDIR_PAYMENT_MIGRATIONS = mkdir -p $(MIGRATIONS_PAYMENT_PATH)

	# Booking service
	MKDIR_BOOKING_GEN = mkdir -p gen/proto/booking
    MKDIR_BOOKING_MIGRATIONS = mkdir -p $(MIGRATIONS_BOOKING_PATH)
endif

run-compose:
	docker-compose up -d --build

stop-compose:
	docker-compose down

# make gen-user
gen-user:
	$(MKDIR_USER_GEN)
	protoc --proto_path=proto/user --go_out=gen/proto/user --go_opt=paths=source_relative --go-grpc_out=gen/proto/user --go-grpc_opt=paths=source_relative user.proto

# make migrate-user name=create_users_table
migrate-user:
	$(MKDIR_USER_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

# make gen-flight
gen-flight:
	$(MKDIR_FLIGHT_GEN)
	protoc --proto_path=proto/flight --go_out=gen/proto/flight --go_opt=paths=source_relative --go-grpc_out=gen/proto/flight --go-grpc_opt=paths=source_relative flight.proto

# make migrate-flight name=create_flights_table
migrate-flight:
	$(MKDIR_FLIGHT_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_FLIGHT_PATH) -seq $(name)

# make gen-payment
gen-payment:
	$(MKDIR_PAYMENT_GEN)
	protoc --proto_path=proto/payment --go_out=gen/proto/payment --go_opt=paths=source_relative --go-grpc_out=gen/proto/payment --go-grpc_opt=paths=source_relative payment.proto

# make migrate-payment name=init_schema
migrate-payment:
	$(MKDIR_PAYMENT_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_PAYMENT_PATH) -seq $(name)

# make gen-booking
gen-booking:
	$(MKDIR_BOOKING_GEN)
	protoc --proto_path=proto/booking --go_out=gen/proto/booking --go_opt=paths=source_relative --go-grpc_out=gen/proto/booking --go-grpc_opt=paths=source_relative booking.proto

# make migrate-booking name=init_booking
migrate-booking:
	$(MKDIR_BOOKING_MIGRATIONS)
	migrate create -ext sql -dir $(MIGRATIONS_BOOKING_PATH) -seq $(name)