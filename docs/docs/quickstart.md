# Quickstart

To launch a brand new instance:

```bash
alpine launch #launches with default parameters
alpine launch -a aarch64 #create an aarch64 instance
alpine launch -d 10G -c 4 -m 2048 #create a machine with a 10GB disk, 4 cpus and 2GB of RAM

```

Access instance via ssh:

```bash
alpine launch -s 22 #launch a instance and expose SSH port to host port 22
ssh root@localhost -p 22 #password: root
```

Expose additional instance ports to host:

```bash
alpine launch -s 23 -p 8888,5432 #launch a instance, expose SSH to host port 23 and forward instance ports 8888 and 5432 to host ports 8888 and 5432
```

Instances can be easily packaged for export and re-use as tar.gz files:

```bash
alpine list

NAME                STATUS      SSH    PORTS ARCH        PID
forthright-hook     Running     23           aarch64     91598
hot-cow             Running     22           x86_64      82361
```

```bash
alpine publish hot-cow
```

This will create a file hot-cow.tar.gz which can be imported as:

```bash
alpine import hot-cow.tar.gz
```
