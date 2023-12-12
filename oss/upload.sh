#!/bin/bash

services=(root@47.94.135.134)
base_path=/data/osshelper
service_name=osshe
function upload_service() {
  for ((i = 0; i < ${#services[@]}; i++)); do
    echo "upload service $1 to ${services[i]}"
    ssh ${services[i]} "rm -rf $base_path/$1"
    scp $1 ${services[i]}:$base_path/$1
    ssh ${services[i]} "cd $base_path && chmod +x $1"
  done
}
function build() {
  echo "build service $1"
  GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o $1 .
}

function main() {
  #if  [ ! -n "$1" ] ;then
  #  		echo "请输入程序名"
  #  		exit 1
  #fi
  build $service_name
  upload_service $service_name
  rm $service_name
}

main $service_name
