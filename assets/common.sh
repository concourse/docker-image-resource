permit_device_control() {
  local devices_mount_info=$(cat /proc/self/cgroup | grep devices)

  if [ -z "$devices_mount_info" ]; then
    # cgroups not set up; must not be in a container
    return
  fi

  local devices_subsytems=$(echo $devices_mount_info | cut -d: -f2)
  local devices_subdir=$(echo $devices_mount_info | cut -d: -f3)

  if [ "$devices_subdir" = "/" ]; then
    # we're in the root devices cgroup; must not be in a container
    return
  fi

  cgroup_dir=/tmp/devices-cgroup

  if [ ! -e ${cgroup_dir} ]; then
    # mount our container's devices subsystem somewhere
    mkdir ${cgroup_dir}
  fi

  if ! mountpoint -q ${cgroup_dir}; then
    if ! mount -t cgroup -o $devices_subsytems none ${cgroup_dir}; then
      return 1
    fi
  fi

  # permit our cgroup to do everything with all devices
  # ignore failure in case something has already done this; echo appears to
  # return EINVAL, possibly because devices this affects are already in use
  echo a > ${cgroup_dir}${devices_subdir}/devices.allow || true
}

ensure_loopback() {
  [ -b /dev/loop$1 ] || mknod -m 0660 /dev/loop$1 b 7 $1
}

make_and_setup() {
  ensure_loopback $1
  losetup -f $2
}

setup_graph() {
  set -x

  if ! permit_device_control; then
    echo "could not permit loopback device usage"
    return 1
  fi

  mkdir -p /var/lib/docker

  image=$(mktemp /tmp/docker.img.XXXXXXXX)
  dd if=/dev/zero of=${image} bs=1 count=0 seek=100G
  mkfs.ext4 -F ${image}

  i=0
  until make_and_setup $i $image >/tmp/setup_loopback.log 2>&1; do
    if grep 'No such file or directory' /tmp/setup_loopback.log; then
      i=$(expr $i + 1)
    else
      echo "failed to setup loopback device:"
      cat /tmp/setup_loopback.log
      return 1
    fi
  done

  # ensure extra loopbacks are available for devmapper driver
  for extra in $(seq 10); do
    ensure_loopback $(expr $i + $extra)
  done

  lo=$(losetup -a | grep ${image} | cut -d: -f1)
  if [ -z "$lo" ]; then
    echo "could not locate loopback device"
    return 1
  fi

  mount ${lo} /var/lib/docker

  set +x
}

sanitize_cgroups() {
  mkdir -p /sys/fs/cgroup
  mountpoint -q /sys/fs/cgroup || \
    mount -t tmpfs -o uid=0,gid=0,mode=0755 cgroup /sys/fs/cgroup

  mount -o remount,rw /sys/fs/cgroup

  for sys in `sed -e '1d;s/\([^\t]\)\t.*$/\1/' /proc/cgroups`; do
    grouping=$(cat /proc/self/cgroup | cut -d: -f2 | grep "\\<$sys\\>")
    mountpoint="/sys/fs/cgroup/$grouping"

    mkdir -p "$mountpoint"

    # clear out existing mount to make sure new one is read-write
    if mountpoint -q "$mountpoint"; then
      umount "$mountpoint"
    fi

    mount -n -t cgroup -o "$grouping" cgroup "$mountpoint"

    if [ "$grouping" != "$sys" ]; then
      ln -sf "$mountpoint" "/sys/fs/cgroup/$sys"
    fi
  done
}

start_docker() {
  mkdir -p /var/log
  mkdir -p /var/run

  sanitize_cgroups

  if ! setup_graph >/tmp/setup_graph.log 2>&1; then
    echo "failed to set up graph:"
    cat /tmp/setup_graph.log
    exit 1
  fi

  # check for /proc/sys being remounted readonly, as systemd does
  if mountpoint -q /proc/sys; then
    # remove bind-mounted /proc/sys to restore read-write
    umount /proc/sys
  fi

  local server_args=""

  for registry in $1; do
    server_args="${server_args} --insecure-registry ${registry}"
  done

  docker daemon ${server_args} >/tmp/docker.log 2>&1 &
  echo $! > /tmp/docker.pid

  trap stop_docker EXIT

  sleep 1

  until docker info >/dev/null 2>&1; do
    echo waiting for docker to come up...
    sleep 1
  done
}

stop_docker() {
  local pid=$(cat /tmp/docker.pid)
  if [ -z "$pid" ]; then
    return 0
  fi

  kill -TERM $pid
  wait $pid

  umount /var/lib/docker
}

private_registry() {
  local repository="${1}"

  if echo "${repository}" | fgrep -q '/' ; then
    local registry="$(extract_registry "${repository}")"
    if echo "${registry}" | fgrep -q '.' ; then
      return 0
    fi
  fi

  return 1
}

extract_registry() {
  local repository="${1}"

  echo "${repository}" | cut -d/ -f1
}

extract_repository() {
  local long_repository="${1}"

  echo "${long_repository}" | cut -d/ -f2-
}

image_from_tag() {
  docker images --no-trunc "$1" | awk "{if (\$2 == \"$2\") print \$3}"
}

image_from_digest() {
  docker images --no-trunc --digests "$1" | awk "{if (\$3 == \"$2\") print \$4}"
}

docker_pull() {
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  NC='\033[0m' # No Color

  pull_attempt=1
  max_attempts=3
  while [ "$pull_attempt" -le "$max_attempts" ]; do
    printf "Pulling ${GREEN}%s${NC}" "$1"

    if [ "$pull_attempt" != "1" ]; then
      printf " (attempt %s of %s)" "$pull_attempt" "$max_attempts"
    fi

    printf "...\n"

    if docker pull "$1"; then
      printf "\nSuccessfully pulled ${GREEN}%s${NC}.\n\n" "$1"
      return
    fi

    echo

    pull_attempt=$(expr "$pull_attempt" + 1)
  done

  printf "\n${RED}Failed to pull image %s.${NC}" "$1"
  exit 1
}
