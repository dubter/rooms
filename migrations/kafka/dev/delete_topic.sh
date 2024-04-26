#!/bin/bash
docker exec kafka-0 /opt/bitnami/kafka/bin/kafka-topics.sh --delete --bootstrap-server localhost:9092 --topic messages
