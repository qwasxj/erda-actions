package bin

const PublicExecuteScript = `cd "$(dirname "${BASH_SOURCE[0]}")"
set -o errexit -o nounset -o pipefail

## dir to storage installing package of erda
rm -rf package
mkdir package

## ERDA_VERSION validate
if ! env | grep ERDA_VERSION > /dev/null 2>&1; then
    echo "no specify env ERDA_VERSION"
fi
if [[ -z "$ERDA_VERSION" ]]; then
    echo "ERDA_VERSION is empty"
    exit
fi

## erda actions and addons
rm -rf ./erda/scripts/erda-actions
cp -a /tmp/"$ERDA_VERSION"/erda-actions ./erda/scripts/

rm -rf ./erda/scripts/erda-addons
cp -a /tmp/"$ERDA_VERSION"/erda-addons ./erda/scripts/

CURRENT_PATH="$PWD"
VERSION_PATH="$CURRENT_PATH"/version
ERDA_YAML="$CURRENT_PATH/erda/templates/erda/erda.yaml"

## compose erda.yaml
if [ -f "$ERDA_YAML" ]; then
    rm -rf "$ERDA_YAML"
fi
cd "$VERSION_PATH" &&
./compose.sh "$ERDA_VERSION"
cp -rf erda.yaml "$ERDA_YAML"
cd "$CURRENT_PATH" && rm -rf "$VERSION_PATH"


## mv to action to do
wget https://raw.githubusercontent.com/erda-project/erda/release/${ERDA_VERSION}/docs/guides/deploy/How-to-install-the-Erda.md -O ./erda/How-to-install-the-Erda.md
`
