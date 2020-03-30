#!/bin/sh

BaseDir=$(cd `dirname ${BASH_SOURCE}` ; pwd)
COMMONSS=$BaseDir"/commonss.sh"

source $COMMONSS

APIPort=`getAPIPort $CfgFile`
if [ $? != 0 ]; then
    echo "get apiPort failed"
    exit 1
fi

ssctrlRequest POST "127.0.0.1:${APIPort}/mode" $1
if [ $? != 0 ]; then
    echo "post setmode command failed"
    exit 1
fi

if [ $Code != "200" ]; then
    echo "setmode command error"
    echo "message  : $Message"
    echo "http code: $Code"
fi