include docker-compose/.env

ifeq ($(db), pg)
	url := $(POSTGRES_URL)
else ifeq ($(db), cassandra)
	url := $(CASSANDRA_URL)
endif

docker-local:
	docker compose -f docker-compose/docker-compose-local.yaml --env-file=docker-compose/.env up --remove-orphans --build

docker-dev:
	docker compose -f docker-compose/docker-compose-dev.yaml --env-file=docker-compose/.env up --remove-orphans --build

migrate-create:
	migrate create -ext sql -dir migrations/$(db) -seq $(name)

# Example: > db=pg make migrate-up
migrate-up:
	migrate -path migrations/$(db) -database $(url) up

migrate-down:
	migrate -path migrations/$(db) -database $(url) down

create-kafka-topic-local:
	./migrations/kafka/local/create_topic.sh

delete-kafka-topic-local:
	./migrations/kafka/local/delete_topic.sh

create-kafka-topic-dev:
	./migrations/kafka/dev/create_topic.sh

delete-kafka-topic-dev:
	./migrations/kafka/dev/delete_topic.sh