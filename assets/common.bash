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

  local server_args="${1}"

  local repository=$(jq -r '.source.repository // ""' < $payload)
  if private_registry "${repository}" ; then
    local registry="$(extract_registry "${repository}")"

    server_args="${server_args} --insecure-registry ${registry}"
  fi

  docker ${server_args} -d >/dev/null 2>&1 &

  sleep 1

  until docker info >/dev/null 2>&1; do
    echo waiting for docker to come up...
    sleep 1
  done
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

find_protocol() {
  local registry=$1
  
  https=`curl -k -I https://${registry}/v1/_ping 2>/dev/null | head -n 1 | cut -d$' ' -f2`
  if [[ $https == 200 ]]; then
    echo "https"
  else
    http=`curl -k -I http://${registry}/v1/_ping 2>/dev/null | head -n 1 | cut -d$' ' -f2`
    if [[ $http == 200 ]]; then
      echo "http"
    else
      echo "Failed to ping registry"
      exit 1
    fi
  fi
}

extract_registry() {
  local repository="${1}"

  echo "${repository}" | cut -d/ -f1
}

extract_repository() {
  local long_repository="${1}"

  echo "${long_repository}" | cut -d/ -f2-
}

docker_image() {
  docker images --no-trunc "$1" | awk "{if (\$2 == \"$2\") print \$3}"
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
