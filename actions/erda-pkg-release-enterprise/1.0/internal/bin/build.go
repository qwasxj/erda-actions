package bin

const (
	PrivateExecuteScript = `#!/bin/bash
set -o errexit -o nounset -o pipefail

## 判断是否为版本
function is_old_version() {
    is="0"
    # mac perl grep
    if echo "$1" | grep -E '^\d+.\d+$' >/dev/null 2>/dev/null; then
        is="1"
    fi
    # linux perl grep
    if echo "$1" | grep -P '^\d+.\d+$' >/dev/null 2>/dev/null; then
        if [ "$is" = "0" ]; then
            is="1"
        fi
    fi
    echo "$is"
}

## 获取 开源前 version 的所有版本号
function get_old_versions() {
    versions=()
    for dir in $(ls "$1"); do
        if [ $(is_old_version "$dir") = "1" ]; then
            versions+=("$dir")
        fi
    done
    echo "${versions[*]}"
}

## 判断是否为版本
function is_version() {
    is="0"
    # mac perl grep
    if echo "$1" | grep -E '^[v|V]*\d+.\d+[\-rc]*' >/dev/null 2>/dev/null; then
        is="1"
    fi
    # linux perl grep
    if echo "$1" | grep -P '^[v|V]*\d+.\d+[\-rc]*' >/dev/null 2>/dev/null; then
        if [ "$is" = "0" ]; then
            is="1"
        fi
    fi
    echo "$is"
}

## get main.beta version of erda
function get_main_beta_version() {
    local main_beta_version="$1"

    ## filter v in version
    if echo "$main_beta_version" | grep "v" > /dev/null 2>&1; then
        main_beta_version="${main_beta_version#*v}"
    fi

    ## filter -rc in version
    if echo "$main_beta_version" | grep "\-rc" > /dev/null 2>&1; then
        main_beta_version="${main_beta_version%-rc*}"
    fi

    ## filter patch version in version
    dotNum=$(echo "$main_beta_version" | awk -F"." '{print NF-1}')
    if [ "$dotNum" == "2" ]; then
        main_beta_version="${main_beta_version%.*}"
    fi

    echo "$main_beta_version"
}

function check_env() {
    if ! printenv | grep $1>/dev/null 2>/dev/null; then
        echo "you do not set $1 var"
        exit 1
    fi
}

## 参数个数， target-dice-version 校验
if [ $# -lt 1 ]; then
cat << USAGE
usage:
    $0 pack
USAGE
exit 1
fi

check_env DICE_VERSION

if [ $(is_version "$DICE_VERSION") != "1" ]; then
    echo "dice target version: $DICE_VERSION invalid"
    exit 1
fi

MAIN_BETA_VERSION=$(get_main_beta_version "$DICE_VERSION")

echo "registry.cn-hangzhou.aliyuncs.com/terminus/koperator:${MAIN_BETA_VERSION}" > ./offline/dice.txt

## 处理 migration job
./bin/init_migrate

## build tools
make clean
args=$1
buildos=$(uname | tr 'A-Z' 'a-z')
make $args os=linux buildos=$buildos

## 处理 version extensions
CURRENT_PATH="$PWD"
TEMP="$PWD"/build_temp
DICE_VERSION_PATH="$TEMP"/version

mkdir -p "$TEMP" && cd "$TEMP"
cp -a "$REPO_VERSION_PATH" ./

## 4.0 之后对应发版 version 的处理
if [ "$ERDA_TO_PUBLIC" == "true" ]; then
    cp -a "/tmp/$DICE_VERSION" ./version/
fi

## k8s 平台需生成个组件发布的 dice.yaml, dcos 平台需个组件发布的 dice.yaml
cd "$DICE_VERSION_PATH" &&
./compose.sh "$DICE_VERSION"

cd "$CURRENT_PATH" &&
cp "$DICE_VERSION_PATH"/dice.yaml  ./dice-tools/versionpackage
cp "$DICE_VERSION_PATH"/fdp.yaml  ./dice-tools/versionpackage

## releases 在 dcos 里面有用(后面平台迁移完后干掉)
cp -r "$DICE_VERSION_PATH"/"$DICE_VERSION"/releases ./dice-tools/versionpackage
if [ ! -d ./dice-tools/versionpackage/extensions ]; then
    mkdir -p ./dice-tools/versionpackage/extensions
fi

## 4.0 及之前 addon、actions 处理； 否则 archive action 归档到 oss
## erda-pkg-release-enterprise 准备好了 extensions 到 /tmp/patch_version 版本目录
if [ "$ERDA_TO_PUBLIC" == "false" ]; then
    mkdir -p /tmp/"$DICE_VERSION"/extensions
    cd /tmp/"$DICE_VERSION"/extensions &&
    git clone https://"${GIT_ACCOUNT}":"${GIT_TOKEN}"@github.com/erda-project/erda-actions.git
    git clone https://"${GIT_ACCOUNT}":"${GIT_TOKEN}"@github.com/erda-project/erda-actions-enterprise.git
    git clone https://"${GIT_ACCOUNT}":"${GIT_TOKEN}"@github.com/erda-project/erda-addons.git
    git clone https://"${GIT_ACCOUNT}":"${GIT_TOKEN}"@github.com/erda-project/erda-addons-enterprise.git
    cd "$CURRENT_PATH" && echo $(pwd)
fi

cp -r /tmp/"$DICE_VERSION"/extensions/erda-actions/actions ./dice-tools/versionpackage/extensions
cp -r /tmp/"$DICE_VERSION"/extensions/erda-actions-enterprise/actions/* ./dice-tools/versionpackage/extensions/actions/
cp -r /tmp/"$DICE_VERSION"/extensions/erda-addons/addons ./dice-tools/versionpackage/extensions
cp -r /tmp/"$DICE_VERSION"/extensions/erda-addons-enterprise/addons/* ./dice-tools/versionpackage/extensions/addons

## 下载 dice cli
curl -o ./dice-tools/bin/dice http://terminus-dice.oss.aliyuncs.com/installer/dice-${MAIN_BETA_VERSION}
chmod 755 ./dice-tools/bin/dice

## 下载 dcos 安装工具，以后干掉
curl -o ./dice-tools/scripts/dcosctl/operator http://terminus-dice.oss.aliyuncs.com/installer/operator-${MAIN_BETA_VERSION}
chmod 755 ./dice-tools/scripts/dcosctl/operator
echo "$MAIN_BETA_VERSION" > ./dice-tools/scripts/dcosctl/VERSION

## copy migration sql，contains all 3.9 - 4.0 migrate sql and version specified by the DICE_VERSION
echo "$DICE_VERSION" > ./dice-tools/versionpackage/version
versions=($(get_old_versions "$DICE_VERSION_PATH"))
for version in "${versions[@]}"; do
    if  [ -d "$DICE_VERSION_PATH"/"$version"/migrations ]; then
        mkdir -pv ./dice-tools/versionpackage/"$version"
        cp -r "$DICE_VERSION_PATH"/"$version"/migrations/* ./dice-tools/versionpackage/"$version"
    fi
done
## version after erda become public
if [ ! -d /dice-tools/versionpackage/"$DICE_VERSION" ]; then
    ## version exists
    if [[ -d "$DICE_VERSION_PATH/$DICE_VERSION/migrations" || -d "$DICE_VERSION_PATH/$DICE_VERSION/sqls" ]]; then
        mkdir -p ./dice-tools/versionpackage/"$DICE_VERSION"
        if [ -d "$DICE_VERSION_PATH/$DICE_VERSION/migrations" ]; then
            cp -r "$DICE_VERSION_PATH/$DICE_VERSION/migrations"/* ./dice-tools/versionpackage/"$DICE_VERSION"/
        fi
        if [ -d "$DICE_VERSION_PATH/$DICE_VERSION/sqls" ]; then
            for dir in $(ls "$DICE_VERSION_PATH"/"$DICE_VERSION"/sqls); do
                if [ -d "$DICE_VERSION_PATH"/"$DICE_VERSION"/sqls/"$dir" ]; then
                    mkdir -pv ./dice-tools/versionpackage/"$DICE_VERSION"/"$dir"
                    cp "$DICE_VERSION_PATH"/"$DICE_VERSION"/sqls/"$dir"/* ./dice-tools/versionpackage/"$DICE_VERSION"/"$dir"/
                fi
                if [ -f "$DICE_VERSION_PATH"/"$DICE_VERSION"/sqls/"$dir" ]; then
                    cp "$DICE_VERSION_PATH"/"$DICE_VERSION"/sqls/"$dir" ./dice-tools/versionpackage/"$DICE_VERSION"
                fi
            done
        fi
    ## version does not exist
    else
        echo "version does not exists"
        exit 1
    fi
fi

## 拷贝 init sql
cp -r "$DICE_VERSION_PATH/3.9/init/sqls" ./dice-tools/versionpackage

if [[ -f "$DICE_VERSION_PATH"/dice.yaml ]]; then
    rm "$DICE_VERSION_PATH"/dice.yaml
fi
rm -rf "$TEMP"

cat ./dice-tools/versionpackage/dice.yaml | grep '^\s*image:' | sed 's/^\s*image:\s*//' | tr -d '"' | tr -d "'" | sort -u >> ./offline/dice.txt
find ./dice-tools/versionpackage/extensions -name dice.yml -exec grep -F image: {} \; | sed -e 's/^\s*image:\s*//' -e 's/^\s*DEFAULT_DEP_CACHE_IMAGE:\s*//' | sed '/^\s*$/d' | sort -u > ./offline/ext.txt

echo "build tools successfully..."

echo
echo 'To Build Offline:'
echo '    bash offline/build.sh'
echo`
)
