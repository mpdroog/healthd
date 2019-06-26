#!/bin/bash
# Ensure all $procs are running
procs=(cron nginx)
err=0
for proc in ${procs[@]}; do
  if pgrep $proc > /dev/null
  then
    :
  else
    echo "CRITICAL - No such proc=${proc}"
    err=1
  fi
done

if [ $err = 1 ]; then
  exit 1
else
  echo "OK"
fi