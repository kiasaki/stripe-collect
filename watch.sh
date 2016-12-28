#!/bin/sh

# Needs `notifywait` from https://github.com/ggreer/fsevents-tools on OSX

function finish {
  killall stripe-collect
  printf "\E(B\E[m"
}
trap finish EXIT

while true; do
  printf "\E[32m"
  make build
  printf "\E(B\E[m"

  ./stripe-collect &
  PID=$!

  printf "\E[36m"
  notifywait views
  printf "\E(B\E[m"

  kill $PID

  while curl --output /dev/null --silent --head --fail http://localhost:3000; do
    sleep 0.1
  done
done
