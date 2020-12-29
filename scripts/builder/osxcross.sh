#!/bin/sh

set -eu

OSXCROSS_REPO=tpoechtrager/osxcross
OSXCROSS_SHA1=c2ad5e8
OSX_SDK=10.15
OSX_SDK_SUM=b6ba29a1219593dee432563c493b1c0db2c9788b99bfa7fdf329ef37e02e11b7

# darwin
mkdir -p /usr/x86_64-apple-darwin/osxcross
mkdir -p /tmp/osxcross && cd "/tmp/osxcross"
curl -sLo osxcross.tar.gz "https://codeload.github.com/${OSXCROSS_REPO}/tar.gz/${OSXCROSS_SHA1}"
tar --strip=1 -xzf osxcross.tar.gz
rm -f osxcross.tar.gz
curl -sLo tarballs/${OSX_SDK}.tar.xz "https://github.com/phracker/MacOSX-SDKs/releases/download/${OSX_SDK}/MacOSX${OSX_SDK}.sdk.tar.xz"
echo "${OSX_SDK_SUM}"  "tarballs/${OSX_SDK}.tar.xz" | sha256sum -c -
yes "" | SDK_VERSION=10.15 OSX_VERSION_MIN=10.15 OCDEBUG=1 ./build.sh
mv target/* /usr/x86_64-apple-darwin/osxcross/
mv tools /usr/x86_64-apple-darwin/osxcross/
cd /usr/x86_64-apple-darwin/osxcross/include
ln -s ../SDK/MacOSX10.15.sdk/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/CarbonCore.framework/Versions/A/Headers/ CarbonCore
ln -s ../SDK/MacOSX10.15.sdk/System/Library/Frameworks/CoreFoundation.framework/Versions/A/Headers/ CoreFoundation
ln -s ../SDK/MacOSX10.15.sdk/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/ Frameworks
ln -s ../SDK/MacOSX10.15.sdk/System/Library/Frameworks/Security.framework/Versions/A/Headers/ Security
rm -rf /tmp/osxcross
rm -rf "/usr/x86_64-apple-darwin/osxcross/SDK/MacOSX10.15.sdk/usr/share/man"
# symlink ld64.lld
ln -s /usr/x86_64-apple-darwin/osxcross/bin/x86_64-apple-darwin19-ld /usr/x86_64-apple-darwin/osxcross/bin/ld64.lld
ln -s /usr/x86_64-apple-darwin/osxcross/lib/libxar.so.1 /usr/lib