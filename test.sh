#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"

sudo docker build -f test-files/Dockerfile -t openvpn-setup-test .

if [ "${1}" == "-i" ]; then
    exec sudo docker run -v `pwd`/:/openvpn-setup -v `pwd`/test-files/run.sh:/run.sh -i -t openvpn-setup-test /bin/bash
else
    time sudo docker run -v `pwd`/:/openvpn-setup -i -t openvpn-setup-test /openvpn-setup/test-files/run.sh
    if [ $? -ne 0 ]; then
        echo "Tests failed."
        exit $?
    else
        echo "Tests passed."
    fi
fi
