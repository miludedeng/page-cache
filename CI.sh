#!/bin/bash
# package source
gox --osarch=linux/amd64
docker rm `docker ps -a | grep 'page-cache' | col 2`
docker rmi `docker images | grep "^page-cache" | col 3`
docker build -t page-cache .

function col {
  awk -v col=$1 '{print $col}'
}
