#!/bin/bash

go run /openvpn-setup/app.go setup-server -c /openvpn-setup/example-server-config.json

go run /openvpn-setup/app.go client -n test1

cd /etc/openvpn
openvpn --config /etc/openvpn/server.conf
