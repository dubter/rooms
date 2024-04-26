#!/bin/bash
docker exec kafka-0 /opt/bitnami/kafka/bin/kafka-topics.sh --create --bootstrap-server localhost:9092 --topic messages --partitions 6 --replication-factor 3