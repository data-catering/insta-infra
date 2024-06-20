#!/usr/bin/env bash

count=0
total=0

for f in $(ls /tmp/data/*.cql);
do
  total=$((total + 1))
  cqlsh -u "${CASSANDRA_USER:-cassandra}" -p "${CASSANDRA_PASSWORD:-cassandra}" -f "${f}" cassandra 9042
  if [[ $? -eq 0 ]];
  then
    count=$((count + 1))
  fi;
done;

echo "Executed ${count} out of ${total} files"