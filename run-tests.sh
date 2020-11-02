#!/bin/bash

set -e

init_rpc() {
    echo "* Starting rpcbind"
    if [ ! -x /run/rpcbind ] ; then
        install -m755 -g 32 -o 32 -d /run/rpcbind
    fi
    rpcbind || return 0
    rpc.statd -L || return 0
    rpc.idmapd || return 0
}

init_dbus() {
    echo "* Starting dbus"
    if [ ! -x /var/run/dbus ] ; then
        install -m755 -g 81 -o 81 -d /var/run/dbus
    fi
    rm -f /var/run/dbus/*
    rm -f /var/run/messagebus.pid
    dbus-uuidgen --ensure
    dbus-daemon --system --fork
}

init_rpc
sleep 0.5
init_dbus
sleep 0.5
ganesha.nfsd -L STDERR -F -f /etc/ganesha/ganesha.conf &
sleep 3

set +e
echo "Running NFS tests!"
/app/runtests test localhost:2049 /
success=$?
if [ $success -eq 0 ]; then
  echo "::set-output success=true"
else
  echo "::set-output success=false"
fi
exit $success
