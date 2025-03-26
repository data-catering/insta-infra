#!/usr/bin/env bash

host="kafka"
# blocks until kafka is reachable
kafka-topics --bootstrap-server ${host}:29092 --list

if [[ -z "$KAFKA_TOPICS" ]]; then
  echo "No Kafka topics provided via KAFKA_TOPICS environment variable"
  exit 0
fi

echo "Creating kafka topics"
for t in $(echo "$KAFKA_TOPICS" | sed s/,/\\n/g); do
  kafka-topics --create --if-not-exists --topic "${t}" --bootstrap-server ${host}:29092 --replication-factor 1 --partitions 1
done

echo "Successfully created the following topics:"
kafka-topics --bootstrap-server ${host}:29092 --list
