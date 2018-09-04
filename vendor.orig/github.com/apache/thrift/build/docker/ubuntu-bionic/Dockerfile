# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Apache Thrift Docker build environment for Ubuntu Bionic
# Using all stock Ubuntu Bionic packaging except for:
# - d: dmd does not come with Ubuntu
# - dart: does not come with Ubuntu
# - dotnet: does not come with Ubuntu
# - go: want latest
# - nodejs: want v8, bionic comes with v6
# - openssl: to support dlang and the deimos for openssl, need to use 1.0 not 1.1
#

FROM buildpack-deps:bionic-scm
MAINTAINER Apache Thrift <dev@thrift.apache.org>
ENV DEBIAN_FRONTEND noninteractive

### Add apt repos

RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y --no-install-recommends \
      apt \
      apt-transport-https \
      apt-utils \
      curl \
      dirmngr \
      software-properties-common \
      wget

# csharp (mono) - if we ever want a later version
# RUN echo "deb http://download.mono-project.com/repo/debian xenial main" | tee /etc/apt/sources.list.d/mono.list && \
#     apt-key adv --keyserver keyserver.ubuntu.com --recv-keys A6A19B38D3D831EF

# Dart
RUN curl https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    curl https://storage.googleapis.com/download.dartlang.org/linux/debian/dart_stable.list > \
      /etc/apt/sources.list.d/dart_stable.list

# dotnet (netcore)
RUN curl https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > /etc/apt/trusted.gpg.d/microsoft.gpg && \
    wget -q -O /etc/apt/sources.list.d/microsoft-prod.list https://packages.microsoft.com/config/ubuntu/18.04/prod.list && \
    chown root:root /etc/apt/trusted.gpg.d/microsoft.gpg && \
    chown root:root /etc/apt/sources.list.d/microsoft-prod.list

# node.js
RUN curl -sL https://deb.nodesource.com/gpgkey/nodesource.gpg.key | apt-key add - && \
    echo "deb https://deb.nodesource.com/node_8.x bionic main" | tee /etc/apt/sources.list.d/nodesource.list

### install general dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
`# General dependencies` \
      bash-completion \
      bison \
      build-essential \
      clang \
      cmake \
      debhelper \
      flex \
      gdb \
      llvm \
      ninja-build \
      pkg-config \
      valgrind \
      vim
ENV PATH /usr/lib/llvm-6.0/bin:$PATH

# boost-1.62 has a terrible bug in boost::test, see https://svn.boost.org/trac10/ticket/12507
RUN apt-get install -y --no-install-recommends \
`# C++ dependencies` \
      libboost-all-dev \
      libevent-dev \
      libssl-dev \
      qt5-default \
      qtbase5-dev \
      qtbase5-dev-tools

RUN apt-get install -y --no-install-recommends \
`# csharp (mono) dependencies` \
      mono-devel

ENV SBCL_VERSION 1.4.8
RUN \
`# Common Lisp (sbcl) dependencies` \
    curl --version && \
    curl -O -J -L https://kent.dl.sourceforge.net/project/sbcl/sbcl/${SBCL_VERSION}/sbcl-${SBCL_VERSION}-x86-64-linux-binary.tar.bz2 && \
    tar xjf sbcl-${SBCL_VERSION}-x86-64-linux-binary.tar.bz2 && \
    cd sbcl-${SBCL_VERSION}-x86-64-linux && \
    ./install.sh && \
    sbcl --version && \
    cd .. && \
    rm -rf sbcl*

ENV D_VERSION     2.080.0
ENV DMD_DEB       dmd_2.080.0-0_amd64.deb
RUN \
`# D dependencies` \
    wget -q http://downloads.dlang.org/releases/2.x/${D_VERSION}/${DMD_DEB} && \
    dpkg --install ${DMD_DEB} && \
    rm -f ${DMD_DEB} && \
    mkdir -p /usr/include/dmd/druntime/import/deimos /usr/include/dmd/druntime/import/C && \
    curl -sSL https://github.com/D-Programming-Deimos/libevent/archive/master.tar.gz| tar xz && \
    mv libevent-master/deimos/* /usr/include/dmd/druntime/import/deimos/ && \
    mv libevent-master/C/* /usr/include/dmd/druntime/import/C/ && \
    rm -rf libevent-master && \
    curl -sSL https://github.com/jeking3/openssl/archive/tls_method.tar.gz| tar xz && \
    mv openssl-tls_method/deimos/* /usr/include/dmd/druntime/import/deimos/ && \
    mv openssl-tls_method/C/* /usr/include/dmd/druntime/import/C/ && \
    rm -rf openssl-tls_method

RUN apt-get install -y --no-install-recommends \
      `# Dart dependencies` \
      dart/stable
ENV PATH /usr/lib/dart/bin:$PATH

RUN apt-get install -y --no-install-recommends \
`# dotnet core dependencies` \
      dotnet-sdk-2.1

RUN apt-get install -y --no-install-recommends \
`# Erlang dependencies` \
      erlang-base \
      erlang-eunit \
      erlang-dev \
      erlang-tools \
      rebar

RUN apt-get install -y --no-install-recommends \
`# GlibC dependencies` \
      libglib2.0-dev

# golang
ENV GOLANG_VERSION 1.10.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 4b677d698c65370afa33757b6954ade60347aaca310ea92a63ed717d7cb0c2ff
RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz && \
      echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - && \
            tar -C /usr/local -xzf golang.tar.gz && \
                  ln -s /usr/local/go/bin/go /usr/local/bin && \
                        rm golang.tar.gz

RUN apt-get install -y --no-install-recommends \
`# Haskell dependencies` \
      ghc \
      cabal-install

RUN apt-get install -y --no-install-recommends \
`# Haxe dependencies` \
      haxe \
      neko \
      neko-dev && \
    haxelib setup --always /usr/share/haxe/lib && \
    haxelib install --always hxcpp 2>&1 > /dev/null

RUN apt-get install -y --no-install-recommends \
`# Java dependencies` \
      ant \
      ant-optional \
      openjdk-8-jdk \
      maven && \
    update-alternatives --set java /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java

RUN apt-get install -y --no-install-recommends \
`# Lua dependencies` \
      lua5.2 \
      lua5.2-dev
# https://bugs.launchpad.net/ubuntu/+source/lua5.3/+bug/1707212
# lua5.3 does not install alternatives!
# need to update our luasocket code, lua doesn't have luaL_openlib any more

RUN apt-get install -y --no-install-recommends \
`# Node.js dependencies` \
      nodejs

RUN apt-get install -y --no-install-recommends \
`# OCaml dependencies` \
      ocaml \
      opam && \
    opam init --yes && \
    opam install --yes oasis

RUN apt-get install -y --no-install-recommends \
`# Perl dependencies` \
      libbit-vector-perl \
      libclass-accessor-class-perl \
      libcrypt-ssleay-perl \
      libio-socket-ssl-perl \
      libnet-ssleay-perl

RUN apt-get install -y --no-install-recommends \
`# Php dependencies` \
      php \
      php-cli \
      php-dev \
      php-pear \
      re2c \
      composer

RUN apt-get install -y --no-install-recommends \
`# Python dependencies` \
      python-all \
      python-all-dbg \
      python-all-dev \
      python-ipaddress \
      python-pip \
      python-setuptools \
      python-six \
      python-tornado \
      python-twisted \
      python-wheel \
      python-zope.interface && \
   pip install --upgrade backports.ssl_match_hostname

RUN apt-get install -y --no-install-recommends \
`# Python3 dependencies` \
      python3-all \
      python3-all-dbg \
      python3-all-dev \
      python3-pip \
      python3-setuptools \
      python3-six \
      python3-tornado \
      python3-twisted \
      python3-wheel \
      python3-zope.interface

RUN apt-get install -y --no-install-recommends \
`# Ruby dependencies` \
      ruby \
      ruby-dev \
      ruby-bundler

RUN apt-get install -y --no-install-recommends \
`# Rust dependencies` \
      cargo \
      rustc

# cppcheck-1.82 has a nasty cpp parser bug, so we're using something newer
RUN apt-get install -y --no-install-recommends \
`# Static Code Analysis dependencies` \
      cppcheck \
      sloccount && \
    pip install flake8 && \
    wget -q "https://launchpad.net/ubuntu/+source/cppcheck/1.83-2/+build/14874703/+files/cppcheck_1.83-2_amd64.deb" && \
    dpkg -i cppcheck_1.83-2_amd64.deb && \
    rm cppcheck_1.83-2_amd64.deb

# Clean up
RUN rm -rf /var/cache/apt/* && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/*

ENV THRIFT_ROOT /thrift
RUN mkdir -p $THRIFT_ROOT/src
COPY Dockerfile $THRIFT_ROOT/
WORKDIR $THRIFT_ROOT/src
