---
date: "2021-06-28T00:00:00+00:00"
title: "Installation"
---

Grizzly is currently available for Linux and MacOS systems.

Download the [latest release](https://github.com/grafana/grizzly/releases).

Select and download an appropriate file for your operating system. Then:
```
sudo mv $DOWNLOADED_FILE /usr/local/bin/grr
sudo chmod +x /usr/local/bin/grr
```
If you wish to build the latest (as yet unreleased) version, assuming you have
a recent Golang installed:

```
git clone https://github.com/grafana/grizzly.git
cd grizzly
make dev
sudo mv grr /usr/local/bin/grr
```