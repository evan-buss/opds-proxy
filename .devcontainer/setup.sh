#!/bin/bash

wget https://github.com/pgaskin/kepubify/releases/download/v4.0.4/kepubify-linux-64bit
mv kepubify-linux-64bit /usr/local/bin/kepubify
chmod +x /usr/local/bin/kepubify
rm -rf kepubify-linux-64bit


wget https://web.archive.org/web/20150803131026if_/https://kindlegen.s3.amazonaws.com/kindlegen_linux_2.6_i386_v2_9.tar.gz
mkdir kindlegen
tar xvf kindlegen_linux_2.6_i386_v2_9.tar.gz --directory kindlegen
cp kindlegen/kindlegen /usr/local/bin/kindlegen
chmod +x /usr/local/bin/kindlegen 
rm -rf kindlegen kindlegen_linux_2.6_i386_v2_9.tar.gz

go install github.com/air-verse/air@latest