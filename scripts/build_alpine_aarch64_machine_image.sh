VERSION=3.16
MINORVERSION=0
ARCH="aarch64"
ISO=https://dl-cdn.alpinelinux.org/alpine/v$VERSION/releases/$ARCH/alpine-standard-$VERSION.$MINORVERSION-$ARCH.iso


wget $ISO


# inside the vm run alpine-setup
# Allow root ssh login: yes
# Which disk you would like to use: sda
# How would you like to use it: sys