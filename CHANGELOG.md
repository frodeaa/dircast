## 0.4.0 (2016-05-21)

Features:

 - adds a `--log` to enable log if HTTP request

Bugs:

 - fix trailing nulls read from ID3 tag
 - fix Makefile, build would fail if GOPATH was initial empty

## 0.3.0 (2015-03-29)

Features:

 - adds a `--auto-image` to use ID3 attached image as RSS Channel image

Bugs:

 - Fix item URL path, the path would contain `//`, e.g. http://server:port//media.mp3

## 0.2.0 (2015-02-24)

Features:

  - adds a `--bind` flag to run a HTTP server hosting the RSS feed
  - adds a `--file` parameter to support custome file types (default is mp3)
