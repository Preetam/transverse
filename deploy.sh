#!/bin/sh

usage() {
	echo "Usage: ./build.sh <sha>"
}

if [ "$1" == "" ]; then
	echo "Missing argument."
	echo
	usage
	exit 1
fi

echo Running: ansible-playbook -e "build_sha=$1" ./ansible/deploy.yml -i ./ansible/hosts

ansible-playbook -e "build_sha=$1" ./ansible/deploy.yml -i ./ansible/hosts
