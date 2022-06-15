# Installation

Macpine depends on QEMU >= 6.2.0_1:

```bash
#brew update
#brew upgrade
brew install qemu
```

## Install the latest binary

Download the [latest binary release](https://github.com/beringresearch/macpine/releases) for your system and add it to your path by placing to e.g. `/usr/local/bin/`

```bash
wget https://github.com/beringresearch/macpine/releases/download/v.05/alpine
sudo mv alpine /usr/local/bin/
```

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make darwin
```