#!/bin/sh
touch crontab.tmp
echo "*/${CLEAN_INTERVAL_MINUTES:-60} * * * * /app/miniflux-yt-cleaner" > crontab.tmp
crontab crontab.tmp
rm -rf crontab.tmp
exec "$@"
