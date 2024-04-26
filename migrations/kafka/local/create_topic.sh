#!/bin/bash
docker exec kafka-local /opt/bitnami/kafka/bin/kafka-topics.sh --create --bootstrap-server localhost:9092 --topic messages --partitions 6