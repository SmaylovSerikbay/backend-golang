#!/bin/sh

host=$(echo $1 | cut -d: -f1)
port=$(echo $1 | cut -d: -f2)
shift

until nc -z $host $port; do
  >&2 echo "Service is unavailable - sleeping"
  sleep 1
done

>&2 echo "Service is up - executing command"
exec "$@" 