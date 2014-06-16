function start_docker() {
  if [ -f /var/log/docker.pid ]; then
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

  docker -d >>/var/log/docker.out.log 2>>/var/log/docker.err.log &
  docker_pid=$!

  echo $docker_pid > /var/run/docker.pid

  until docker info >/dev/null 2>&1; do
    echo waiting for docker to come up...
    sleep 0.5

    if ! kill -0 $docker_pid >/dev/null 2>&1; then
      echo "docker exited."

      if [ -f /var/log/docker.out.log ]; then
        echo
        echo "stdout:"
        cat /var/log/docker.out.log
      fi

      if [ -f /var/log/docker.err.log ]; then
        echo
        echo "stderr:"
        cat /var/log/docker.err.log
      fi

      exit 1
    fi
  done
}

function docker_image() {
  docker images "$1" | awk "{if (\$2 == \"$2\") print \$3}"
}
