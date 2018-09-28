#!/bin/sh

#set -x

echo "Call the application generating the supervisord's conf file"
/opt/supervisord/bin/bootstrap

echo "Copy supervisord files to their target location"
cp -r /opt/supervisord /var/lib/
