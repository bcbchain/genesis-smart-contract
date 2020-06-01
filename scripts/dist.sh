#!/bin/bash
set -e

contractList="governance smartcontract organization token-basic token-issue netgovernance black-list ibc mining"
TEMP_DIR=./temp/
DIST_DIR=./build/dist/

# Get the version from the environment, or try to figure it out.
if [ -z "$VERSION" ]; then
	VERSION=$(awk -F\" '/ContractSemVer =/ { print $2; exit }' < version/version.go)
fi
if [ -z "$VERSION" ]; then
    echo "Please specify a version."
    exit 1
fi
VERSION="v$VERSION"
echo "==> Building version $VERSION..."

# Delete the old dir
echo "==> Removing old directory..."
rm -rf "$DIST_DIR"

# prepare dist direction
mkdir -p "$TEMP_DIR"
#####################################################################
#                dist & sign contract                               #
#####################################################################

echo "==> Dist & Sign contract..."
for contractName in $contractList
do
  echo "$contractName"
  contractDir="$contractName"
  versionLine=$(grep -r "@:version:" "$contractDir")
  ver=$(echo "$versionLine" | cut -d: -f 4)
  contractTar="$TEMP_DIR$contractName-$ver".tar.gz

  ./scripts/smcpack -n dev@genesis -p aB1@cD2# -s "$contractDir" -o "$TEMP_DIR" >/dev/null
  ./scripts/sigorg -n dev@genesis -p aB1@cD2# -s "$contractTar" >/dev/null

done

#####################################################################
#                         create genesis file                       #
#####################################################################

echo "==> dist genesis file..."
mkdir -p "$DIST_DIR"
for chainID in local devtest
do
  ./scripts/genesis -i "$chainID" -c ./charters/"$chainID" -t "$TEMP_DIR" -o "$DIST_DIR"/"$chainID"/v2 -p Ab1@Cd3$
  cp -rf "$TEMP_DIR"/* "$DIST_DIR"/"$chainID"/v2
done

#####################################################################
#                       dist genesis                                #
#####################################################################
genesisDir="./genesis/"
smcRunSvr="genesis-smcrunsvc_"
smcContract="genesis-smart-contract_"

pushd "$DIST_DIR" >/dev/null
tar -zcf "../$smcContract$VERSION.tar.gz" ./*
mv "../$smcContract$VERSION.tar.gz" "./$smcContract$VERSION.tar.gz"
rm -rf "devtest"
rm -rf "local"
popd >/dev/null

mkdir -p "$genesisDir""temp/genesis/src/"
cp -r "$genesisDir""cmd" "$genesisDir""genesis"  "$genesisDir""stubcommon" "$genesisDir""temp/genesis/src/"
pushd "$genesisDir""temp/" >/dev/null

tar -zcf "../../$DIST_DIR$smcRunSvr$VERSION".tar.gz "genesis"
popd >/dev/null

rm -rf "$TEMP_DIR" "$genesisDir""temp"

# Make the checksums.
pushd "$DIST_DIR" > /dev/null
shasum -a256 ./* > contract_"$VERSION"_SHA256SUMS
popd >/dev/null

exit 0