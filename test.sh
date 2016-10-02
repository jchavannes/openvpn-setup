#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"

go build app.go

sudo docker build -f test-files/Dockerfile -t openvpn-setup-test .

if [ "${1}" == "-i" ]; then
    exec sudo docker run -v `pwd`/:/openvpn-setup -v `pwd`/test-files/app.sh:/app -i -t openvpn-setup-test /bin/bash
else
    sudo docker run -v `pwd`/app:/app -i -t openvpn-setup-test /app test
fi
