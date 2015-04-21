# dircast [![Build Status](https://travis-ci.org/frodeaa/dircast.svg)](https://travis-ci.org/frodeaa/dircast)

Simple tool to generate RSS feed
from mp3 files found in a directory

    usage: dircast [<flags>] <directory>

    Flags:
      --help             Show help.
      -s, --server=http://localhost:8000/
                         hostname (and path) to the root e.g. http://myserver.com/rss
      -b, --bind         Start HTTP server, bind to the server
      --log              Enable log of HTTP requests
      -r, --recursive    how to handle the directory scan
      -a, --auto-image   Resolve RSS image automatically, will use ID3 attached
                         image, image overrides this option, only available in
                         combination with bind
      -l, --language=LANGUAGE
                         the language of the RSS document, a ISO 639 value
      -t, --title="RSS FEED"
                         RSS channel title
      -d, --description=DESCRIPTION
                         RSS channel description
      -i, --image=IMAGE  Image URL for the RSS channel image
      -f, --file="mp3"   File type to include in the RSS document
      --version          Show application version.

    Args:
      <directory>  directory to read files relative from

## Build and test

    make
    bin/dircast --help

## Install

    make
    sudo make install
