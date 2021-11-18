#!/bin/bash

# Usage: 
# `is-service-up.sh [service name] [sleeping time in seconds] [grep search string to check]
# where only $1 is mandatory, the rest optional
# e. g. `is-service-up.sh some.service 60 ' active'`
# Note that this script cannot return any output, which will be ignored by
# systemctl.
# (gwyneth 20211118)

# Extract parameters, if they exist
# At least, we need the service name; abort otherwise with error 22 (invalid
# argument) (gwyneth 20211118)
if [ $# -eq 0 ]; then exit 22; fi

# echo "Arguments: $1 $2"
# init variables with command-line arguments
serviceName="$1"
sleepTime="$2"

if [ -z $sleepTime ]
then
	sleepTime=60
fi

# echo "Service: $serviceName $sleepTime"

# First, check that the service is active and enabled.
# is-enabled is not totally quiet (even with --quiet) so we redirect stderr
# to /dev/null
function is_service_active {
#	echo "Inside function call, launching command"
	/usr/bin/systemctl --quiet is-active $serviceName | /usr/bin/systemctl --quiet is-enabled $serviceName > /dev/null 2>&1
#	/usr/bin/systemctl status $serviceName | /usr/bin/grep $searchString > /dev/null 2>&1
	local retvalue=$?
#	echo "Inside function call, result was: "$retvalue
	return $retvalue
}


# Debugging... (one day, it will be a command option...)
#is_service_active
#echo "Debug run: "$?
#if [ $? -eq 0 ]
#then
#	echo "Active, status was $?"
#else
#	echo "Dead, status was $?"
#fi

# Loop until service becomes active; sleep a bit between checks
is_service_active
until [ $? -eq 0 ]
do
#	echo "Got $?" 
	sleep $sleepTime
	is_service_active
done