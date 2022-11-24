#!/bin/sh

redis-server --daemonize yes --logfile "redis-server.log" --loglevel notice &

until [ "$(redis-cli ping)" == "PONG" ]; do
    sleep 1
done

exec /go/bin/dmon "$@"
