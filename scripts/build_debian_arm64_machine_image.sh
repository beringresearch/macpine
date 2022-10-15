qemu-img create -f qcow2 debian-3607-aarch64.qcow2 32G 

wget http://ftp.au.debian.org/debian/dists/bullseye/main/installer-arm64/current/images/netboot/debian-installer/arm64/initrd.gz

wget http://ftp.au.debian.org/debian/dists/bullseye/main/installer-arm64/current/images/netboot/debian-installer/arm64/linux

wget http://ftp.au.debian.org/debian/dists/bullseye/main/installer-arm64/current/images/netboot/mini.iso

qemu-system-aarch64 -M virt -cpu cortex-a53 -m 1G -kernel ./linux -initrd ./initrd.gz \
    -hda debian-3607-aarch64.qcow2 -append "console=ttyAMA0" \
    -drive file=mini.iso,id=cdrom,if=none,media=cdrom \
    -device virtio-scsi-device -device scsi-cd,drive=cdrom -nographic