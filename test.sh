#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"

sudo docker build -f test-files/Dockerfile -t openvpn-setup-test .

RUN_COMMAND="sudo docker run --cap-add=NET_ADMIN"
COMMON_MOUNTS="-v /dev/net/tun:/dev/net/tun -v `pwd`/:/openvpn-setup -v /dev/random:/dev/random -v /dev/urandom:/dev/urandom"
COMMON_PORTS="-p 1194:1194 -p 1194:1194/udp"

if [ "${1}" == "-i" ]; then
    exec ${RUN_COMMAND} ${COMMON_MOUNTS} ${COMMON_PORTS} -v `pwd`/test-files/run.sh:/run.sh -i -t openvpn-setup-test /bin/bash
else
    time ${RUN_COMMAND} ${COMMON_MOUNTS} ${COMMON_PORTS} -i -t openvpn-setup-test /openvpn-setup/test-files/run.sh
    if [ $? -ne 0 ]; then
        echo "Tests failed."
        exit $?
    else
        echo "Tests passed."
    fi
fi
