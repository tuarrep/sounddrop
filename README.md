# Sounddrop : Multiroom audio system
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop?ref=badge_shield)

## Overview

Sound Drop is a golang software to play music in multiple rooms at the same time.
It's designed to run on multiple devices and allows them to discover themselves automatically on the local network.
User can after create group of devices to share sound between them.

## Basic usage
**Versions of go prior to 1.8 aren't supported**
```bash
go get ./...

# on a device with sound files
go run main.go -auto-accept -auto-start-stream -playlist-dir=/path/to/sounds/folder

# on the others devices (on the same network)
go run main.go 
```

## Author

Nicolas Perraut - [@mafzst](https://github.com/mafzst)


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop?ref=badge_large)

## Third party softwares
See [FOSSA Report](https://app.fossa.io/attribution/f77bb2ca-2a14-4cc0-98c9-eb438f6814fe)
