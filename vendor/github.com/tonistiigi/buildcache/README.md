### buildcache

Store Docker buildcache to a separate file that can be reapplied after `docker pull`.

#### install

`go get github.com/tonistiigi/buildcache/cmd/buildcache`

#### usage

After building your image run.

```
buildcache save -o cache.tgz imagename
```

This will create a small file `cache.tgz` that contains all known build cache for the image.

When you want to restore it in another machine run:

```
docker pull imagename
docker load -i cache.tgz
```

#### on problems

Build cache can only be applied if the pulled image is the same that was built. Easy way to check that is to check if the ID of the image that was first built is the same that you got after `docker pull` in another machine.

If `docker load` succeeds but cache still isn't being used in another machine try running `docker history id-of-pulled-image` and `docker history id-of-built-image` and compare the results to see where you got a cache miss.

#### compatibility

Buildcache works with Docker v1.12 using only the remote API. To use buildcache in Docker v1.11 it needs to access the Docker storage directory directly. Use `-g` options to specify directory other than `/var/lib/docker`. Eariler Docker versions are not supported.
