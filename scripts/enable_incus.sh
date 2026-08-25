#!/bin/sh
set -e

hwclock -s

> /etc/apk/repositories
echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

apk update
apk upgrade
apk add --no-cache zfs incus incus-client ip6tables lxc linux-pam shadow-uidmap

# Make sure config directories exist before writing into them
mkdir -p /etc/lxc /etc/pam.d /etc/conf.d /var/lib/incus

> /etc/init.d/incusd
cat > /etc/init.d/incusd <<'endmsg'
#!/sbin/openrc-run
command="/usr/sbin/incusd"
command_args="${INCUSD_OPTIONS}"
command_background="true"
pidfile="/run/incus/${RC_SVCNAME}.pid"
retry="${INCUSD_STOP_TIMEOUT:-60}"
extra_started_commands="quit"
description_quit="Daemon quits and leaves the instances running"
: ${INCUSD_FORCE_STOP:="no"}
depend() {
        need net cgroups dbus
        use lxcfs
        after firewall
}
start_pre() {
        checkpath --directory "${pidfile%/*}" --mode 0750
}
stop() {
        ebegin "Stopping ${RC_SVCNAME}"
        if [ "$INCUSD_FORCE_STOP" = "no" ]; then
                $command shutdown --timeout ${INCUSD_STOP_TIMEOUT:-60}
        elif [ "$INCUSD_FORCE_STOP" = "yes" ]; then
                $command shutdown --force
        fi
}
quit() {
        ebegin "Quitting ${RC_SVCNAME}"
        start-stop-daemon --signal SIGQUIT --pidfile $pidfile --quiet
        rm /run/openrc/started/incusd
}
endmsg

chmod +x /etc/init.d/incusd

echo "session optional pam_cgfs.so -c freezer,memory,name=systemd,unified" >> /etc/pam.d/system-login
echo "lxc.idmap = u 0 100000 65536" >> /etc/lxc/default.conf
echo "lxc.idmap = g 0 100000 65536" >> /etc/lxc/default.conf
echo "root:100000:65536" >> /etc/subuid
echo "root:100000:65536" >> /etc/subgid
echo "systemd_container=yes" >> /etc/conf.d/lxc

chmod u+s /usr/bin/newuidmap /usr/bin/newgidmap

# Ensure DNS resolution works before incusd starts — it fetches instance-type
# data from images.linuxcontainers.org on first run and needs working DNS.
if ! grep -q '^nameserver' /etc/resolv.conf 2>/dev/null; then
    echo "nameserver 1.1.1.1" >> /etc/resolv.conf
fi

echo "Waiting for network connectivity..."
for i in $(seq 1 30); do
    if ping -c 1 -W 2 1.1.1.1 >/dev/null 2>&1; then
        echo "Network is up."
        break
    fi
    sleep 1
done

rc-update add incusd default
rc-service incusd start

reboot