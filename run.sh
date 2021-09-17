#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=9001 &
./server -port=9002 &
./server -port=9004 -api=1 &

wait