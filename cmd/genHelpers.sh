#!/bin/bash -e

IN="$1"
OUT="$2"

if test "${IN}x" = "x"; then
  echo "Please provide input file (hint helper.ts)";
  exit 1
fi

if test "${OUT}x" = "x"; then
  echo "Please provide output file (hint helpers_gen.go)";
  exit 1
fi

if test ! -e "${IN}"; then
  echo "${IN} doesn't exist";
  exit 1
fi

INJS="${IN/\.ts/\.js}"

echo "Creating ${OUT}"

tsc -t es6 -m es6 "${IN}" || exit 1

echo package struct2ts > "${OUT}"
echo >> "${OUT}"
echo "const ts_helpers = \`" >> "${OUT}"
cat "${IN}" >> "${OUT}"
echo "\`" >> "${OUT}"

echo >> "${OUT}"
echo "const es6_helpers = \`" >> "${OUT}"
perl -pe 's/\s{4}/\t/g' < "$INJS" >> "${OUT}"
rm "${INJS}"
echo "\`" >> "${OUT}"
