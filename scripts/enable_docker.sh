echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

apk update
apk upgrade

apk add --no-cache docker

addgroup root docker

rc-update add docker boot
service docker start

reboot
docker run hello-world