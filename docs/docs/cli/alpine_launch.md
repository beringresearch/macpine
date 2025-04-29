# alpine launch

Create and start an instance.

```
alpine launch
```

## Description

Create and start an instance.

## Options

```
  -a, --arch string     Machine architecture. Defaults to host architecture.
  -c, --cpu string      Number of CPUs to allocate. (default "2")
  -d, --disk string     Disk space (in bytes) to allocate. K, M, G suffixes are supported. (default "10G")
  -h, --help            help for launch
  -i, --image string    Image to be launched. (default "alpine_3.16.0")
  -m, --memory string   Amount of memory (in MB) to allocate. (default "2048")
      --mount string    Path to a host directory to be shared with the instance.
  -n, --name alpine     Instance name for use in alpine commands.
  -p, --port ,          Forward additional host ports. Multiple ports can be separated by ,.
  -v, --shared          Toggle whether to use mac's native vmnet-shared mode.
  -s, --ssh string      Host port to forward for SSH (required). (default "22")
```
