# Shell completions

You can use command `completion` command to create bash|fish|zsh|powershell completion file

```bash
alex@vosjod:~$ alpine completion
2023/01/05 23:59:48 missing shell
alex@vosjod:~$ alpine completion --help
Generate shell autocompletions. Valid arguments are bash, zsh, and fish.

Usage:
  alpine completion [bash|zsh|fish|powershell]

Flags:
  -h, --help   help for completion
alex@vosjod:~$ alpine completion bash
# bash completion V2 for alpine                               -*- shell-script -*-

__alpine_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Macs have bash3 for which the bash-completion package doesn't include
# _init_completion. This is a minimal version of that function.
__alpine_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

[...]
```


## Install

Create completion file (bash, zsh, fish or powershell) and put on your path, for example using bash:

```bash
alex@vosjod:~$ alpine completion bash > /usr/local/etc/bash_completion.d/alpine
alex@vosjod:~$ source /usr/local/etc/bash_completion.d/alpine
alex@vosjod:~$
```

## Examples

```bash
alex@vosjod:~$ alpine [tab] [tab]
delete   (Delete an instance.)
edit     (Edit instance configuration using Vim.)
exec     (execute COMMAND over ssh.)
help     (Help about any command)
import   (Imports an instance.)
info     (Display information about instances.)
launch   (Launch an Alpine instance.)
list     (List all available instances.)
publish  (Publish an instance.)
ssh      (Attach an interactive shell to an instance.)
start    (Start an instance.)
stop     (Stop an instance.)
alex@vosjod:~$ alpine ssh [tab] [tab]
flat-fight      ignorant-punch
alex@vosjod:~$ alpine ssh flat-fight
2023/01/05 20:30:49 dial tcp [::1]:23: connect: connection refused
```

```bash
alex@vosjod:~$ alpine launch [tab] [tab]
alex@vosjod:~$ alpine launch -[tab] [tab]
--arch    (Machine architecture. Defaults to host cpu architecture.)
--cpu     (Number of CPUs to allocate.)
--disk    (Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix.)
--image   (Image to be launched.)
--memory  (Amount of memory to allocate. Positive integers, in kilobytes.)
--mount   (Path to host directory to be exposed on guest.)
--name    (Name for the instance)
--port    (Forward VM ports to host. Multiple ports can be separated by `,`.)
--ssh     (Forward VM SSH port to host.)
-a        (Machine architecture. Defaults to host cpu architecture.)
-c        (Number of CPUs to allocate.)
-d        (Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix.)
-i        (Image to be launched.)
-m        (Amount of memory to allocate. Positive integers, in kilobytes.)
-n        (Name for the instance)
-p        (Forward VM ports to host. Multiple ports can be separated by `,`.)
-s        (Forward VM SSH port to host.)
alex@vosjod:~$ alpine launch -a aarch64 --[tab] [tab]
--cpu     (Number of CPUs to allocate.)
--disk    (Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix.)
--image   (Image to be launched.)
--memory  (Amount of memory to allocate. Positive integers, in kilobytes.)
--mount   (Path to host directory to be exposed on guest.)
--name    (Name for the instance)
--port    (Forward VM ports to host. Multiple ports can be separated by `,`.)
--ssh     (Forward VM SSH port to host.)
```

