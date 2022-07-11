echo "VERSION_TAG=$VERSION_TAG"

VERSION_REGEX='^v([0-9]+\.[0-9]+\.[0-9]+)(-([0-9]+))?-128tech$'
[[ $VERSION_TAG =~ $VERSION_REGEX ]]

if [ -z $BASH_REMATCH ]; then
    echo "The tagged version does not match the required expression: $VERSION_EXPRESION"
    exit 1
fi

VERSION=${BASH_REMATCH[1]}

if [ ! -z ${BASH_REMATCH[3]} ]; then
    VERSION_PATCH=${BASH_REMATCH[3]}
else
    VERSION_PATCH=1
fi

echo "VERSION=${VERSION}" >> $GITHUB_ENV
echo "VERSION_PATCH=${VERSION_PATCH}" >> $GITHUB_ENV
echo "RPM_NAME=telegraf-128tech-${VERSION}-${VERSION_PATCH}.x86_64.rpm" >> $GITHUB_ENV
