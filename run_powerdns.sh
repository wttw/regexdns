#!/bin/bash

/opt/homebrew/opt/pdns/sbin/pdns_server --no-config --daemon=no --local-port=5300 --launch=pipe --socket-dir=/tmp/pdns --zone-cache-refresh-interval=0 --pipe-command="./regexdns --config testdata/sausagemail.zone"
