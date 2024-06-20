#!/usr/bin/env bash

count=0
total=0

for f in $(ls /tmp/data/*.sql);
do
  total=$((total + 1))
  mysql -u root -p"${MYSQL_PASSWORD:-root}" -h "mysql" < "${f}"
  if [[ $? -eq 0 ]];
  then
    count=$((count + 1))
  fi;
done;

echo "Executed ${count} out of ${total} files"