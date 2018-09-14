# Git Complete Clone
This tool will download all repositories for a git user to the expected location within your `$GOPATH`.

## Prerequisites
- git
- go

## Installation
```
go get github.com/alexdcox/gitcc
```

## Running

### Fetch all repositories
```
gitcc alexdcox
```

### Fetch only repositories for a given language
```
gitcc alexdcox -l go
```