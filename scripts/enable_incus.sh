hwclock -s
> /etc/apk/repositories

echo "http://dl-cdn.alpinelinux.org/alpine/v3.18/main" >> /etc/apk/repositories
echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

apk update
apk upgrade

apk add --no-cache zfs incus incus-client ip6tables


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
        # Required for running systemd containers
        #if [ -d /sys/fs/cgroup/unified ] && ! [ -d /sys/fs/cgroup/systemd ]; then
        #       checkpath --directory --owner root:lxd /sys/fs/cgroup/systemd
        #       mount -t cgroup \
        #               -o rw,nosuid,nodev,noexec,relatime,none,name=systemd \
        #               cgroup /sys/fs/cgroup/systemd
        #fi
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

echo "session optional pam_cgfs.so -c freezer,memory,name=systemd,unified" >> /etc/pam.d/system-login
echo "lxc.idmap = u 0 100000 65536" >> /etc/lxc/default.conf
echo "lxc.idmap = g 0 100000 65536" >> /etc/lxc/default.conf
echo "root:100000:65536" >> /etc/subuid
echo "root:100000:65536" >> /etc/subgid

# If you plan to run systemd based Linux distributions (Debian, Ubuntu, etc.)
echo "systemd_container=yes" >>  /etc/conf.d/lxc


# Sort out UID mappings
chmod -x /usr/bin/newgidmap

rc-update add lxc
rc-update add incusd

reboot