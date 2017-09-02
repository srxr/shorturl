# shorturl
[![GoDoc](https://godoc.org/github.com/prologic/shorturl?status.svg)](https://godoc.org/github.com/prologic/shorturl)
[![Go Report Card](https://goreportcard.com/badge/github.com/prologic/shorturl)](https://goreportcard.com/report/github.com/prologic/shorturl)

shorturl is a web app that allows you to create short urls of much longer more
complex urls for easier sharing or embedding.

## Installation

### Source

Due to the necessity of the way assets are handled if you are building/installing from source and intend to run outside of the source tree you need to do something like this:

```#!bash
$ go get github.com/GeertJohan/go.rice/rice
$ go get github.com/prologic/shorturl
$ cd $GOPATH/src/github.com/prologic/shorturl
$ rice embed-go
$ go build
```

## Usage

Run shorturl:

```#!bash
$ shorturl
```

Then visit: http://localhost:8000/

## Configuration

By default shorturl stores urls in `urls.db` in the local directory. This can
be configured with the `-dbpath /path/to/urls.db` option.

shorturl also displays an absolute url after creating and uses the value of
`-baseurl` (*default: `""`*) for display purposes. This is useful for copying
and pasting the shorturl.

## License

MIT
