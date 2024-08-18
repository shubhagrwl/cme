#!/usr/bin/env bash

echo "Waiting for Cassandra to be ready..."
until printf "" 2>>/dev/null >>/dev/tcp/cassandra/9042; do 
    sleep 5;
    echo "Waiting for Cassandra...";
done

echo "Cassandra is up. Creating keyspace..."
cqlsh cassandra -u admin -p admin -e "CREATE KEYSPACE IF NOT EXISTS your_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};"

echo "Keyspace creation script finished."
