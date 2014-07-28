function start_docker() {
  if /etc/init.d/docker.io status | grep "docker is running"; then
    return 0
  fi

  mkdir -p /var/log
  mkdir -p /var/run

  # set up cgroups
  mkdir -p /sys/fs/cgroup
  cgroups-mount

  # make /dev/shm larger
  mount -t tmpfs -o remount,size=1G none /dev/shm

  # docker graph dir
  mkdir -p /var/lib/docker
  mount -t tmpfs -o size=1G none /var/lib/docker

  /etc/init.d/docker.io start

  until docker info >/dev/null 2>&1; do
    echo waiting for docker to come up...
    sleep 0.5
  done
}

function docker_image() {
  docker images "$1" | awk "{if (\$2 == \"$2\") print \$3}"
}
