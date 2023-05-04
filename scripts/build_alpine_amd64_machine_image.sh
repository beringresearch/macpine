qemu-img create -f qcow2 alpine_3.16.0_lxd-amd64.qcow2 32G 

wget https://dl-cdn.alpinelinux.org/alpine/v3.16/releases/x86_64/alpine-standard-3.16.0-x86_64.iso

qemu-system-x86_64 -cpu qemu64 -m 1G \
    -hda alpine_3.16.0_lxd-amd64.qcow2 \
    -cdrom alpine-standard-3.16.0-x86_64.iso \
    -boot d \
    -smp 2 \
    -nic user,hostfwd=tcp::4022-:22