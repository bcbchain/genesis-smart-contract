#!/bin/bash

function GetDirs() {
  IDIRS=()
  i=1
  for _ in $(cat ./scripts/.distignore)
  do
    NUM=$i
    IDIR=$(awk 'NR=='$NUM' {print $1}' ./scripts/.distignore)
    if [[ -n "$IDIR" ]]; then
      IDIRS[$i]=$IDIR
    fi

    : $(( i++ ))
  done

  for f in `ls -l $PWD`
  do
    if [[ -d "$f" ]];then
      b=0
      for id in "${IDIRS[@]}"
      do
        if [[ "$id" == "$f" ]];then
          b=1
        fi
      done

      if [[ $b == 0 ]];then
        TDIRS[${#TDIRS[*]}]=$f
      fi
    fi
  done
  return 0
}

DIST_DIR=./build/dist/
TDIRS=()
cd ..

echo "==> Removing old directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

GetDirs

#SRC_DIRS=""
#for d in "${TDIRS[@]}"
#do
#  if [[ -z "$SRC_DIRS" ]];then
#    SRC_DIRS="$d"
#  else
#    SRC_DIRS="$SRC_DIRS"" $d"
#  fi
#
#done

# mkdir temp for contracts tar
TEMP_DIR=./temp/
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

echo "==> Dist & Sign contract..."
for contractName in "${TDIRS[@]}"
do
  echo "==> packing $contractName..."
  contractDir="$contractName"
  versionLine=$(grep -r "@:version:" "$contractDir")
  ver=$(echo "$versionLine" | cut -d: -f 4)
  contractTar="$TEMP_DIR$contractName-$ver".tar.gz

  ./scripts/smcpack -n dev@genesis -p aB1@cD2# -s "$contractDir" -o "$TEMP_DIR" >/dev/null
  ./scripts/sigorg -n dev@genesis -p aB1@cD2# -s "$contractTar" >/dev/null

done

echo "==> Dist genesis json file..."
mkdir -p "$DIST_DIR"
for chainID in local devtest
do
  ./scripts/genesis -i "$chainID" -c ./charters/"$chainID" -t "$TEMP_DIR" -o "$DIST_DIR"/"$chainID"/v2 -p Ab1@Cd3$ >/dev/null
  cp -rf "$TEMP_DIR"/* "$DIST_DIR"/"$chainID"/v2
done

pushd "$DIST_DIR" >/dev/null
tar -zcf "../$project_name""_$VERSION.tar.gz" ./*
mv "../$project_name""_$VERSION.tar.gz" "./$project_name""_$VERSION.tar.gz"
rm -rf "devtest"
rm -rf "local"
popd >/dev/null

rm -rf "$TEMP_DIR"

# Make the checksums.
pushd "$DIST_DIR" > /dev/null
shasum -a256 ./* > "$project_name"_SHA256SUMS
popd >/dev/null

echo ""
echo "==> Build results:"
echo "==> Path: "$DIST_DIR""
echo "==> Files: "
ls -hl "$DIST_DIR"