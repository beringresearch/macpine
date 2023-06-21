qemu-img create -f qcow2 alpine_3.16.0_lxd-x86_64.qcow2 1G 

wget https://dl-cdn.alpinelinux.org/alpine/v3.16/releases/x86_64/alpine-standard-3.16.0-x86_64.iso

qemu-system-x86_64 -cpu qemu64 -m 1G \
    -hda alpine_3.16.0_lxd-x86_64.qcow2 \
    -cdrom alpine-standard-3.16.0-x86_64.iso \
    -boot d \
    -smp 2 \
    -netdev user,id=net0,hostfwd=tcp::3024-:22 \
    -device e1000,netdev=net0,mac=56:c9:81:7f:9c:0b

# inside the vm run alpine-setup
# Allow root ssh login: yes
# Which disk you would like to use: sda
# How would you like to use it: sys