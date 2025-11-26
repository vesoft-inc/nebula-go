# docker file for building the go application
# e.g. docker build -t reg.vesoft-inc.com/vesoft-ng/nebula-ng-golang:nightly .
FROM golang:1.22.4-alpine AS builder

WORKDIR /usr/src/nebula-ng-golang

COPY . .
