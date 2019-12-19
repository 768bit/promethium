#!/usr/bin/env bash

## setup the environment for use - we need firecracker source so we can build it...

KVM_CHECKED=false
KVM_ALLOWED=false

function checkKVM() {
    case $KVM_ALLOWED in
      (true)    return;;
    esac
    echo -n "Checking KVM Access: "
    (
      [ -r /dev/kvm ] && [ -w /dev/kvm ] && KVM_ALLOWED=true || KVM_ALLOWED=false
    )
    case $KVM_ALLOWED in
      (true)    echo "OK" && return;;
      (false)   echo "FAIL";;
    esac
    (
      fixKVM
    )
}

function fixKVM() {
    case $KVM_CHECKED in
      (true)    return;;
      (false)   KVM_CHECKED=true;;
    esac
    [ $(stat -c "%G" /dev/kvm) = kvm ] && sudo usermod -aG kvm ${USER} && echo "KVM Access granted for $USER."
    checkKVM
}


FIRECRACKER_VERSION="v0.20.0"
RUST_TOOLCHAIN="1.39.0"
TMP_BUILD_DIR=/tmp/build

CURR_DIR=`pwd`

SCRIPT_DIR="`dirname "$0"`"

ENV_DIRECTORY="`readlink -f "$SCRIPT_DIR/../"`"

SCRIPT_DIR="$ENV_DIRECTORY/scripts"

WORK_DIR="$ENV_DIRECTORY/workspace"


export CARGO_HOME="$SCRIPT_DIR/support/docker/rust"
export RUSTUP_HOME="$SCRIPT_DIR/support/docker/rust"
mkdir -p "$RUSTUP_HOME"
mkdir -p "$CARGO_HOME"

mkdir "$TMP_BUILD_DIR"
cd "$TMP_BUILD_DIR"
curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain "$RUST_TOOLCHAIN"
$RUSTUP_HOME/bin/rustup target add aarch64-unknown-linux-gnu
$RUSTUP_HOME/bin/rustup target add aarch64-unknown-linux-gnu
$RUSTUP_HOME/bin/rustup component add rustfmt
$RUSTUP_HOME/bin/rustup component add clippy-preview
$RUSTUP_HOME/bin/cargo install -v cargo-kcov
$RUSTUP_HOME/bin/cargo install -v cargo-audit
$RUSTUP_HOME/bin/cargo kcov --print-install-kcov-sh | sh
rm -rf "$TMP_BUILD_DIR"

cd "$SCRIPT_DIR/support/docker"

docker build -f Dockerfile.x86_64 -t promethium/fc-dev-amd64 .
docker build -f Dockerfile.x86_64_cross -t promethium/fc-dev-aarch64-cross .

##exit 0

echo "Creating workspace dir $WORK_DIR"

mkdir "$WORK_DIR"

cd $WORK_DIR

rm -rf "$WORK_DIR/firecracker"

git clone https://github.com/firecracker-microvm/firecracker

cd "$WORK_DIR/firecracker"

echo "Checking out Firecracker Version $FIRECRACKER_VERSION"

git checkout tags/$FIRECRACKER_VERSION

echo "Building Firecracker Release Binary"

mkdir -p "$ENV_DIRECTORY/assets/firecracker/x86_64"
mkdir -p "$ENV_DIRECTORY/assets/firecracker/aarch64"

FC_X64_BUILT_PATH="$WORK_DIR/firecracker/build/cargo_target/x86_64-unknown-linux-musl/release"

FC_ARM64_BUILT_PATH="$WORK_DIR/firecracker/build/cargo_target/aarch64-unknown-linux-musl/release"

cp "$SCRIPT_DIR/support/devtool" tools/devtoolX
chmod +x tools/devtoolX

tools/devtoolX build --release --arch x86_64
cp -v "$FC_X64_BUILT_PATH/firecracker" "$ENV_DIRECTORY/assets/firecracker/x86_64/"
cp -v "$FC_X64_BUILT_PATH/jailer" "$ENV_DIRECTORY/assets/firecracker/x86_64/"

##exit 0

cd $WORK_DIR

rm -rf "$WORK_DIR/firecracker"

git clone https://github.com/firecracker-microvm/firecracker

cd "$WORK_DIR/firecracker"

cp "$SCRIPT_DIR/support/devtool" tools/devtoolX
chmod +x tools/devtoolX

cat >./.cargo/config <<EOF
[build]
target = "aarch64-unknown-linux-gnu"

[target.aarch64-unknown-linux-gnu]
linker = "aarch64-linux-gnu-gcc"
rustflags = [
  "-L", "/usr/xcc/aarch64-unknown-linux-gnueabi/aarch64-unknown-linux-gnueabi/sysroot/usr/lib",
  "-C", "link-arg=-lgcc",
	"-C", "link-arg=-lftd",
]
EOF

#   "-C", "link-arg=--sysroot=/usr/xcc/aarch64-unknown-linux-gnueabi/aarch64-unknown-linux-gnueabi/sysroot",
#   "-C", "link-arg=-L", "-C", "link-arg=/usr/xcc/aarch64-unknown-linux-gnueabi/aarch64-unknown-linux-gnueabi/sysroot/usr/lib",


tools/devtoolX build --release --arch aarch64 -l gnu
cp "$FC_ARM64_BUILT_PATH/firecracker" "$ENV_DIRECTORY/assets/firecracker/aarch64/"
cp "$FC_ARM64_BUILT_PATH/jailer" "$ENV_DIRECTORY/assets/firecracker/aarch64/"

##exit 0

checkKVM

## now we are here  firecracker is built lets make sure we have our deps...

mkdir -p "$ENV_DIRECTORY/assets/osv/x86_64/capstan"
mkdir -p "$ENV_DIRECTORY/assets/osv/aarch64/capstan"


cd $WORK_DIR

git clone https://github.com/cloudius-systems/osv.git

cd "$WORK_DIR/osv"

sudo scripts/setup.py

git submodule update --init --recursive

make -j$(nproc)

scripts/build-capstan-img cloudius/osv-base          httpserver,cloud-init                 "OSv base image for developers"
scripts/build-capstan-img cloudius/osv-node          node,httpserver,cloud-init            "Node/OSv"
scripts/build-capstan-img cloudius/osv               cli,httpserver,cloud-init           "OSv with shell for users"
cp -rfv build/capstan/cloudius/osv $HOME/.capstan/repository/cloudius/
cp build/release/usr.img $HOME/.capstan/repository/cloudius/osv/osv.qemu

cp build/release/usr.img "$ENV_DIRECTORY/assets/osv/x86_64/"
cp build/release/loader-stripped.elf "$ENV_DIRECTORY/assets/osv/x86_64/"

mkdir "$ENV_DIRECTORY/assets/osv/x86_64/capstan"

cp -rfv build/capstan/cloudius/*/*.qemu.gz "$ENV_DIRECTORY/assets/osv/x86_64/capstan/"

cd "$WORK_DIR"

rm -rf "$WORK_DIR/osv"

git clone https://github.com/cloudius-systems/osv.git

cd "$WORK_DIR/osv"

sudo scripts/setup.py

git submodule update --init --recursive

ARCH="aarch64" make -j$(nproc)
