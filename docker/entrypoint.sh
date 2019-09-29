#!/usr/bin/env bash

set -eo pipefail

cd /opt/mcbuild

find /opt/mcserver -maxdepth 0 -empty -exec cp -r ./* /opt/mcserver/ \;

cp /opt/mcbuild/bedrockd /opt/mcserver/
cp /opt/mcbuild/bedrock_server /opt/mcserver/

cd /opt/mcserver

exec ./bedrockd
