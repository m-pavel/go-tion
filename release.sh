#!/bin/bash
IMPL=muka

make clean
rm -rf dst/*

mkdir -p ./dst/386
make IMPL=$IMPL ARCH=386
mv ./tion-* ./dst/386
tar -czvf ./dst/386.tgz -C ./dst/386/ .

mkdir -p ./dst/amd64
make IMPL=$IMPL ARCH=amd64
mv ./tion-* ./dst/amd64
tar -czvf ./dst/amd64.tgz -C ./dst/amd64/ .

mkdir -p ./dst/arm
make IMPL=$IMPL ARCH=arm
mv ./tion-* ./dst/arm
tar -czvf ./dst/arm.tgz -C ./dst/arm/ .


mkdir -p ./dst/arm64
make IMPL=$IMPL ARCH=arm64
mv ./tion-* ./dst/arm64
tar -czvf ./dst/arm64.tgz -C ./dst/arm64/ .


