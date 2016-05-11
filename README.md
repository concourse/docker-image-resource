# Docker Image Resource

Tracks and builds [Docker](https://docker.io) images.

## Source Configuration

* `repository`: *Required.* The name of the repository, e.g.
`concourse/docker-image-resource`.

* `tag`: *Optional.* The tag to track. Defaults to `latest`.

* `username`: *Optional.* The username to authenticate with when pushing.

* `password`: *Optional.* The password to use when authenticating.

* `email`: *Optional.* The email to use when authenticating.

* `insecure_registries`: *Optional.* An array of CIDRs or `host:port` addresses
  to whitelist for insecure access (either `http` or unverified `https`).

* `registry_mirror`: *Optional.* A URL pointing to a docker registry mirror service.

Note: docker registry must be [v2](https://docs.docker.com/registry/spec/api/).

## Behavior

### `check`: Check for new images.

The current image digest is fetched from the registry for the given tag of the
repository.


### `in`: Fetch the image from the registry.

Pulls down the repository image by the requested digest.

The following files will be placed in the destination:

* `/image`: If `save` is `true`, the `docker save`d image will be provided
  here.
* `/repository`: The name of the repository that was fetched.
* `/tag`: The tag of the repository that was fetched.
* `/image-id`: The fetched image ID.
* `/digest`: The fetched image digest.
* `/rootfs.tar`: If `rootfs` is `true`, the contents of the image will be
  provided here.

#### Parameters

* `save`: *Optional.* Place a `docker save`d image in the destination.
* `rootfs`: *Optional.* Place a `.tar` file of the image in the destination.
* `skip_download`: *Optional.* Skip `docker pull` of image. Only `/image-id`,
  `/repository`, and `/tag` will be populated. `/image` and `/rootfs.tar` will
  not be present.


### `out`: Push an image, or build and push a `Dockerfile`.

Push a Docker image to the source's repository and tag. The resulting
version is the image's digest.

#### Parameters

* `rootfs`: *Optional.* Default `false`. Place a `.tar` file of the image in
  the destination.

* `build`: *Optional.* The path of a directory containing a `Dockerfile` to
  build.

* `dockerfile`: *Optional.* The path of the `Dockerfile` in the directory if it's not at the root of the directory.

* `cache`: *Optional.* Default `false`. When the `build` parameter is set,
  first pull `image:tag` from the Docker registry (so as to use cached
  intermediate images when building). This will cause the resource to fail
  if it is set to `true` and the image does not exist yet.

* `load_base`: *Optional.* A path to a directory containing an image to `docker
  load` before running `docker build`. The directory must have `image`,
  `image-id`, `repository`, and `tag` present, i.e. the tree produced by `/in`.

* `load_file`: *Optional.* A path to a file to `docker load` and then push.

* `import_file`: *Optional.* A path to a file to `docker import` and then push.

* `pull_repository`: *Optional.* A path to a repository to pull down, and then
  push to this resource.

* `tag`: *Optional.* The value should be a path to a file containing the name
  of the tag.

* `tag_prefix`: *Optional.* If specified, the tag read from the file will be
  prepended with this string. This is useful for adding `v` in front of version
  numbers.

* `tag_as_latest`: *Optional.*  Default `false`. If true, the pushed image will be tag as latest too and tag will be push.


## Example

``` yaml
resources:
- name: git-resource
  type: git
  source: # ...

- name: git-resource-image
  type: docker-image
  source:
    repository: concourse/git-resource
    email: docker@example.com
    username: username
    password: password

- name: git-resource-rootfs
  type: s3
  source: # ...

jobs:
- name: build-rootfs
  plan:
  - get: git-resource
  - put: git-resource-image
    params: {build: git-resource}
    get_params: {rootfs: true}
  - put: git-resource-rootfs
    params: {file: git-resource-image/rootfs.tar}
```
