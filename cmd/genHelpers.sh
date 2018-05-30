#!/bin/sh

IN=$1
INJS=${IN/\.ts/\.js}
OUT=$2
tsc -t es6 -m es6 $IN || exit 1

echo package struct2ts > $OUT
echo >> $OUT
echo "const ts_helpers = \`" >> $OUT
cat $IN >> $OUT
echo "\`" >> $OUT

echo >> $OUT
echo "const es6_helpers = \`" >> $OUT
perl -pe 's/\s{4}/\t/g' < $INJS >> $OUT
rm $INJS
echo "\`" >> $OUT
