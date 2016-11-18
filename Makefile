GHACCOUNT := johansundell
NAME := piwififix
VERSION := v1.1

include common.mk

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox
