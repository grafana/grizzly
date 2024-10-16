---
date: "2021-06-28T00:00:00+00:00"
title: "Installation"
---

## Grizzly is currently available for Linux and MacOS systems

### Installing Grizzly on Linux

Download the [latest release](https://github.com/grafana/grizzly/releases).

Select and download an appropriate file for your operating system. Then:

```bash
sudo mv $DOWNLOADED_FILE /usr/local/bin/grr
sudo chmod +x /usr/local/bin/grr
```

### Installing Grizzly on macOS via Homebrew

Before you begin
Install [Homebrew](https://brew.sh) on your computer.

Once Homebrew is installed, you can install Grizzly using the following command:

```bash
brew install grizzly
```

### Building from source

If you wish to build the latest (as yet unreleased) version, assuming you have
a recent Golang installed:

```bash
git clone https://github.com/grafana/grizzly.git
cd grizzly
make dev
sudo mv grr /usr/local/bin/grr
```
