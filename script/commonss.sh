#!/bin/sh

CfgFile=~/.ssctrl/config.toml

getAPIPort() {
    configfile=$1
    Key="apiPort"
    APIPort=`sed "/^[[:blank:]]*$Key[[:blank:]]*=[[:blank:]]*/"'!d; s/.*=[[:blank:]]*//' $configfile`
    RES=$?
    if [ $RES != 0 ]; then
        return $RES
    fi

    APIPort=`echo $APIPort | sed 's/"//g'`
    RES=$?

    if [ -z "$APIPort" ]; then
        APIPort=1083
    fi
    echo $APIPort
    return $RES
}

# Message and Code are the return values of ssctrlRequest 
ssctrlRequest() {
    reqCMD=$1
    reqURL=$2
    reqData=""
    if [ -n "$3" ]; then
        reqData="-d $3"
    fi

    #set -x
    returnData=$(curl -s -w %{http_code} -X $reqCMD "$reqURL" $reqData)
    RES=$?
    if [ $RES != 0 ]; then 
        return $RES
    fi

    # $returnData is a string like 
    # '{"enabled":true,"mode":"pac"}200', or only '404'
    Message=`echo $returnData | sed "s/^\(.*\)\([0-9]\{3\}\)$/\1/"`
    Code=`echo $returnData | sed "s/^\(.*\)\([0-9]\{3\}\)$/\2/"`

    return 0
}