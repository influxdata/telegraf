echo "VERSION_TAG=$VERSION_TAG"

VERSION_REGEX='^v([0-9]+\.[0-9]+\.[0-9]+)(-([0-9]+))?-128tech$'
[[ $VERSION_TAG =~ $VERSION_REGEX ]]

if [ -z $BASH_REMATCH ]; then
    echo "The tagged version does not match the required expression: $VERSION_EXPRESION"
    exit 1
fi

echo "::set-env name=VERSION::${BASH_REMATCH[1]}"

if [ ! -z ${BASH_REMATCH[3]} ]; then
    echo "::set-env name=VERSION_PATCH::${BASH_REMATCH[3]}"
else
    echo "::set-env name=VERSION_PATCH::1"
fi
