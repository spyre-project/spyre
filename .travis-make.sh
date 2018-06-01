#!/bin/sh

cd $(dirname $0)
echo "$0: Running $@ ... "
# work around Travis CI heuristics
( while sleep 60; do echo -n '.' ; done) &
trap "kill $!" EXIT
"$@" > .travis-make.log 2>&1
