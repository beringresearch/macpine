echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories
echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

apk update
apk upgrade

apk add --no-cache lxd lxd-client lxcfs dbus zfs
echo "session optional pam_cgfs.so -c freezer,memory,name=systemd,unified" >> /etc/pam.d/system-login
echo "lxc.idmap = u 0 100000 65536" >> /etc/lxc/default.conf
echo "lxc.idmap = g 0 100000 65536" >> /etc/lxc/default.conf
echo "root:100000:65536" >> /etc/subuid
echo "root:100000:65536" >> /etc/subgid

# If you plan to run systemd based Linux distributions (Debian, Ubuntu, etc.)
echo "systemd_container=yes" >>  /etc/conf.d/lxc

rc-update add lxc
rc-update add lxd
rc-update add lxcfs
rc-update add dbus

reboot