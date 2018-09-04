#
# Coverity Scan Travis build script
# To run this interactively, set the environment variables yourself,
# and run this inside a docker container.
#
# Command-Line Arguments
#
# --skipdownload   to skip re-downloading the Coverity Scan build package (large)
#
# Environment Variables (required)
#
# COVERITY_SCAN_NOTIFICATION_EMAIL  - email address to notify
# COVERITY_SCAN_TOKEN               - the Coverity Scan token (should be secure)
#
# Environment Variables (defaulted)
#
# COVERITY_SCAN_BUILD_COMMAND       - defaults to "build/docker/scripts/autotools.sh"
# COVERITY_SCAN_DESCRIPTION         - defaults to TRAVIS_BRANCH or "master" if empty
# COVERITY_SCAN_PROJECT             - defaults to "thrift"

set -ex

COVERITY_SCAN_BUILD_COMMAND=${COVERITY_SCAN_BUILD_COMMAND:-build/docker/scripts/autotools.sh}
COVERITY_SCAN_DESCRIPTION=${COVERITY_SCAN_DESCRIPTION:-${TRAVIS_BRANCH:-master}}
COVERITY_SCAN_PROJECT=${COVERITY_SCAN_PROJECT:-thrift}

# download the coverity scan package

pushd /tmp
if [[ "$1" != "--skipdownload" ]]; then
  rm -rf coverity_tool.tgz cov-analysis*
  wget https://scan.coverity.com/download/linux64 --post-data "token=$COVERITY_SCAN_TOKEN&project=$COVERITY_SCAN_PROJECT" -O coverity_tool.tgz
  tar xzf coverity_tool.tgz
fi
COVBIN=$(echo $(pwd)/cov-analysis*/bin)
export PATH=$COVBIN:$PATH
popd

# build the project with coverity scan

rm -rf cov-int/
cov-build --dir cov-int $COVERITY_SCAN_BUILD_COMMAND
tar cJf cov-int.tar.xz cov-int/
curl --form token="$COVERITY_SCAN_TOKEN" \
     --form email="$COVERITY_SCAN_NOTIFICATION_EMAIL" \
     --form file=@cov-int.tar.xz \
     --form version="$(git describe --tags)" \
     --form description="$COVERITY_SCAN_DESCRIPTION" \
     https://scan.coverity.com/builds?project="$COVERITY_SCAN_PROJECT"

