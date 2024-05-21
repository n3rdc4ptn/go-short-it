#!/bin/bash

go build -ldflags "-X main.Release=1" main.go
