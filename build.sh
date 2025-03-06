#!/bin/bash
go build -buildvcs=false -o GoVersion .
cp ./GoVersion ./GoVersion.app/Contents/MacOS/
rm ./GoVersion
