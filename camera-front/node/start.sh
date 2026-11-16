#!/bin/sh
cd `dirname $0`
pm2-runtime start index.js --name camera-node 