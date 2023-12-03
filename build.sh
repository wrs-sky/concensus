#!/usr/bin/env bash
#RUN_NAME="consensus"

mkdir -p output/bin output/conf output/scripts output/tmp
cp scripts/* output/scripts/
cp conf/* output/conf/
chmod +x output/

go build -x -o output .
