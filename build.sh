#!/bin/bash
if [ $# = 1 ]; then
	echo "building shared name:" $1
	go build -buildmode=c-shared -o "lib"$1.so 
else
	echo "parameter not right"
fi

exit 0

