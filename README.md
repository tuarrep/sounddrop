# Sounddrop : Multiroom audio system
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop?ref=badge_shield)
[![Treeware](https://img.shields.io/treeware/trees/tuarrpe/sounddrop?label=Treeware)](https://plant.treeware.earth/tuarrep/sounddrop)

## Overview

Sound Drop is a golang software to play music in multiple rooms at the same time.
It's designed to run on multiple devices and allows them to discover themselves automatically on the local network.
User can after create group of devices to share sound between them.

## Basic usage
Binaries are available on [releases page](https://github.com/tuarrep/sounddrop/releases/).

__Warnings__
- Versions of go prior to 1.11 aren't supported
- Only wave and mp3 files sampled at 44100Hz are supported

### Usage
```bash
# on a device with sound files
./sounddrop.linux.amd64 -auto-accept -auto-start-stream -playlist-dir=/path/to/sounds/folder

# on the others devices (on the same network)
./sounddrop.linux.amd64
```

### CLI reference
```
  -auto-accept
        Auto accept discovered devices
  -auto-start-stream
        Auto start audio stream
  -playlist-dir string
        Directory containing audio files to play (default ".")
  -port int
        Server port (default 19416)

```

## Author

Nicolas Perraut - [@tuarrep](https://github.com/tuarrep)


## License
This package is [Treeware](https://treeware.earth). If you use it in production, then we ask that you [**buy the world a tree**](https://plant.treeware.earth/{venfor}/{package}) to thank us for our work. By contributing to the Treeware forest you’ll be creating employment for local families and restoring wildlife habitats.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fmafzst%2Fsounddrop?ref=badge_large)

## Third party softwares
See [FOSSA Report](https://app.fossa.io/attribution/f77bb2ca-2a14-4cc0-98c9-eb438f6814fe)
