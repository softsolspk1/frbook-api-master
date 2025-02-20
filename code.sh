#!/bin/bash
EXTRA_ARGS=""
while [ "$1" != "" ]; do
    case $1 in
        -f | --force)           EXTRA_ARGS="$EXTRA_ARGS --force"
                                ;;
    esac
    shift
done
wand9 --input ./code.yaml --api frBook $EXTRA_ARGS