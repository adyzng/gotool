#!/usr/bin/env bash

rm -rf ./map2struct
rm -rf ~/go/bin/map2struct

go build -o map2struct
cp map2struct ~/go/bin/map2struct
