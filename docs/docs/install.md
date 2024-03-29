# Installation

`macpine` depends on QEMU >= 6.2.0_1:

```bash
#brew update
#brew upgrade
brew install qemu
```

## Install the latest binary

Download the [latest binary release](https://github.com/beringresearch/macpine/releases) for your system and add it to your path by placing to e.g. `/usr/local/bin/`

```bash
wget https://github.com/beringresearch/macpine/releases/download/v1.0/alpine_darwin_arm64
mv alpine_darwin_arm64 alpine
sudo chmod +x alpine
sudo mv alpine /usr/local/bin/
```

## Install via Homebrew (recommended)

```bash
brew install macpine
```

## Install via MacPorts
On MacOS, you can install via [MacPorts](https://www.macports.org/):

```bash
sudo port install macpine
```

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make
make install #install to /usr/local by default, may require sudo
```
