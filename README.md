# zkcli

[![Build Status](https://travis-ci.org/let-us-go/zkcli.svg?branch=master)](https://travis-ci.org/let-us-go/zkcli)
[![Go Report Card](https://goreportcard.com/badge/github.com/let-us-go/zkcli)](https://goreportcard.com/report/github.com/let-us-go/zkcli)

A interactive Zookeeper client.

![zkcli](./zkcli.gif)


## Install

### Mac (Homebrew)

```
brew tap let-us-go/zkcli
brew install zkcli
```

### go install

```
go install github.com/wkl7123/zkcli
```

### Build

```
make release-all
```


## Usage

```shell
$ zkcli ls /test
[abc]
```

```shell
$ zkcli
>>> 
>>> help
ls <path>
get <path> <field[/<subField>][/<subField]>
set <path> [<data>]
gf <path> <filePath>
sf <path> <filePath>
create <path> [<data>]
delete <path>
connect <host:port>
addauth <scheme> <auth>
close
exit
>>>
```
