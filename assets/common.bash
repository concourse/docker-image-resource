start_docker() {
  mkdir -p /var/log
  mkdir -p /var/run

  # set up cgroups
  mkdir -p /sys/fs/cgroup
  mountpoint -q /sys/fs/cgroup || \
    mount -t tmpfs -o uid=0,gid=0,mode=0755 cgroup /sys/fs/cgroup

  for d in `sed -e '1d;s/\([^\t]\)\t.*$/\1/' /proc/cgroups`; do
    mkdir -p /sys/fs/cgroup/$d
    mountpoint -q /sys/fs/cgroup/$d || \
      mount -n -t cgroup -o $d cgroup /sys/fs/cgroup/$d
  done

  # docker graph dir
  mkdir -p /var/lib/docker
  mount -t tmpfs -o size=10G none /var/lib/docker

  docker $1 -d >/dev/null 2>&1 &

  sleep 1

  until docker info >/dev/null 2>&1; do
    echo waiting for docker to come up...
    sleep 1
  done
}

docker_image() {
  docker images --no-trunc "$1" | awk "{if (\$2 == \"$2\") print \$3}"
}

docker_pull() {
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  NC='\033[0m' # No Color
  printf "Attempting to pull ${GREEN}%s${NC}...\n" "$1"

  set +e
  pull_attempts=0
  while [ "$pull_attempts" -lt 3 ]; do
    echo
    pull_attempts=$(expr "$pull_attempts" + 1)
    docker pull "$1"

    if [ "$?" -eq "0" ]; then
      printf "\nSuccessfully pulled ${GREEN}%s${NC}.\n\n" "$1"
      set -e
      return
    fi
  done

  printf "\n${RED}Failed to pull image %s." "$1"
  exit 1
}
