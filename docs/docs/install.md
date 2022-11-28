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
wget https://github.com/beringresearch/macpine/releases/download/v.06/alpine
sudo mv alpine /usr/local/bin/
```

## Install via MacPorts
On MacOS, you can install via [MacPorts](https://www.macports.org/):

```bash
sudo port install macpine
```

See more information [here](https://ports.macports.org/port/macpine/).

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make
```