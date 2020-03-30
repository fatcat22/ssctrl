#!/bin/sh

BaseDir=$(cd `dirname ${BASH_SOURCE}` ; pwd)
COMMONSS=$BaseDir"/commonss.sh"

source $COMMONSS

APIPort=`getAPIPort $CfgFile`
if [ $? != 0 ]; then
    echo "get apiPort failed"
    exit 1
fi

ssctrlRequest GET "127.0.0.1:${APIPort}/config"
if [ $? != 0 ]; then
    echo "get status command failed"
    exit 1
fi

if [ $Code != "200" ]; then
    echo "status command error"
    echo "message  : $Message"
    echo "http code: $Code"
else
    echo $Message
fi