function start_docker() {
  if [ -S /var/run/docker.sock ]; then
    return 0
  fi

  # set up cgroups
  mkdir -p /sys/fs/cgroup
  cgroups-mount

  # make /dev/shm larger
  mount -t tmpfs -o remount,size=1G none /dev/shm

  # docker graph dir
  mkdir -p /var/lib/docker
  mount -t tmpfs -o size=1G none /var/lib/docker

  docker -d &

  until [ -S /var/run/docker.sock ]; do
    echo waiting for docker to come up...
    sleep 0.5
  done
}

function docker_image() {
  docker images "$1" | awk "{if (\$2 == \"$2\") print \$3}"
}
