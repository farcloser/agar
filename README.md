# Agar

> * a golang testing framework that generates test audio files and ease testing of audio cli tools
> * [a colloidal extract of algae; used especially in culture media](https://en.wikipedia.org/wiki/Agar)

![logo.jpg](logo.jpg)

## Purpose

Testing audio tools requires... test audio files.
Agar provides helpers and scripts to ease generation of such files,
specifically for the purpose of testing edge cases against broken or otherwise
damaged audio streams (clipping, brick-walled, truncated, upsampled, etc).

Agar additional uses and bundles in [Tigron](https://github.com/containerd/nerdctl/tree/main/mod/tigron),
a golang test framework specifically designed to test binaries as blackboxes with an expressive syntax, pty handling,
and good debugging information.

Agar is primarily serving Farcloser audio libraries testing needs, but could
presumably be used by any other audio tool.

Note that Agar relies heavily on ffmpeg (of course.)

## Usage

TBD. Look at source.

## Development & tests

### Requirements

* golang of course
* make
* ffmpeg and ffprob
* sox
* metaflac

### Initial setup

```bash
make install-dev-tools
```

### Lifecycle

```bash
make lint
make fix
make test
```