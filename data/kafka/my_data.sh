host="kafka"
# blocks until kafka is reachable
kafka-topics --bootstrap-server $host:29092 --list

echo 'Creating kafka topics'
kafka-topics --bootstrap-server $host:29092 --create --if-not-exists --topic account-topic --replication-factor 1 --partitions 1

echo 'Successfully created the following topics:'
kafka-topics --bootstrap-server $host:29092 --list

#kafka-topics --delete --topic account-topic --bootstrap-server localhost:9092
