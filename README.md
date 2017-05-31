# torrentfs

Yet another BitTorrent filesystem based on FUSE.

# Build

```bash
$ git clone https://github.com/ring00/torrentfs.git
$ cd torrentfs && make
```

# Usage

Type `./torrentfs -h` for help message.

```
$ ./torrentfs -h
BitTorrent Filesystem.

Usage:
  torrentfs --mount <mount> --torrent <torrent> [--download <download>]
  torrentfs -h | --help
  torrentfs --version

Options:
  --mount     Mount point.
  --torrent   Directory containing torrents.
  --download  Directory for storing downloaded files.
  -h --help   Show this screen.
  --version   Show version.
```
