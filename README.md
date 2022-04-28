# Docker Image Resource

Tracks and builds [Docker](https://docker.io) images.

Note: docker registry must be [v2](https://docs.docker.com/registry/spec/api/).

## Source Configuration

* `repository`: *Required.* The name of the repository, e.g.
`concourse/docker-image-resource`.

  Note: When configuring a private registry which requires a login, the 
  registry's address must contain at least one '.' e.g. `registry.local` 
  or contain the port (e.g. `registry:443` or `registry:5000`).
  Otherwise docker hub will be used.
  
  Note: When configuring a private registry **using a non-root CA**,
  you must include the port (e.g. :443 or :5000) even though the docker CLI
  does not require it.

* `tag`: *Optional.* The tag to track. Defaults to `latest`.

* `username`: *Optional.* The username to authenticate with when pushing.

* `password`: *Optional.* The password to use when authenticating.

* `aws_access_key_id`: *Optional.* AWS access key to use for acquiring ECR
  credentials.

* `docker_config_json` : *Optional.* The raw `config.json` file used for authenticating with Docker registries. If specified, `username` and `password` parameters will be ignored. You may find this useful if you need to be authenticated against multiple registries (e.g. pushing to a private registry, but you also also need to pull authenticate to pull images from Docker Hub without being rate-limited).

* `aws_secret_access_key`: *Optional.* AWS secret key to use for acquiring ECR
  credentials.

* `aws_session_token`: *Optional.* AWS session token (assumed role) to use for acquiring ECR
  credentials.

* `insecure_registries`: *Optional.* An array of CIDRs or `host:port` addresses
  to whitelist for insecure access (either `http` or unverified `https`).
  This option overrides any entries in `ca_certs` with the same address.

* `registry_mirror`: *Optional.* A URL pointing to a docker registry mirror service.

  Note: `registry_mirror` is ignored if `repository` contains an explicitly-declared
  registry-hostname-prefixed value, such as `my-registry.com/foo/bar`, in which case
  the registry cited in the `repository` value is used instead of the `registry_mirror`.

* `ca_certs`: *Optional.* An array of objects with the following format:

  ```yaml
  ca_certs:
  - domain: example.com:443
    cert: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
  - domain: 10.244.6.2:443
    cert: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
  ```

  Each entry specifies the x509 CA certificate for the trusted docker registry
  residing at the specified domain. This is used to validate the certificate of
  the docker registry when the registry's certificate is signed by a custom
  authority (or itself).

  The domain should match the first component of `repository`, including the
  port. If the registry specified in `repository` does not use a custom cert,
  adding `ca_certs` will break the check script. This option is overridden by
  entries in `insecure_registries` with the same address or a matching CIDR.

* `client_certs`: *Optional.* An array of objects with the following format:

  ```yaml
  client_certs:
  - domain: example.com
    cert: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    key: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
  - domain: 10.244.6.2
    cert: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    key: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
  ```

  Each entry specifies the x509 certificate and key to use for authenticating
  against the docker registry residing at the specified domain. The domain
  should match the first component of `repository`.

 * `max_concurrent_downloads`: *Optional.* Maximum concurrent downloads.

   Limits the number of concurrent download threads.

 * `max_concurrent_uploads`: *Optional.* Maximum concurrent uploads.

   Limits the number of concurrent upload threads.

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
* `/metadata.json`: Collects custom metadata. Contains the container  `env` variables and running `user`.
* `/docker_inspect.json`: Output of the `docker inspect` on `image_id`. Useful if collecting `LABEL` [metadata](https://docs.docker.com/engine/userguide/labels-custom-metadata/) from your image.

#### Parameters

* `save`: *Optional.* Place a `docker save`d image in the destination.
* `rootfs`: *Optional.* Place a `.tar` file of the image in the destination.
* `skip_download`: *Optional.* Skip `docker pull` of image. Artifacts based
  on the image will not be present.

As with all concourse resources, to modify params of the implicit `get` step after each `put` step you may also set these parameters under a `put` `get_params`. For example:

```yaml
put: foo
params: {...}
get_params: {skip_download: true}
```

### `out`: Push an image, or build and push a `Dockerfile`.

Push a Docker image to the source's repository and tag. The resulting
version is the image's digest.

#### Parameters

* `additional_tags`: *Optional.* Path to a file containing a
  whitespace-separated list of tags. The Docker build will additionally be
  pushed with those tags.

* `build`: *Optional.* The path of a directory containing a `Dockerfile` to
  build.

* `build_args`: *Optional.* A map of Docker build-time variables. These will be
  available as environment variables during the Docker build.
  
  While not stored in the image layers, they are stored in image metadata and
  so it is recommend to avoid using these to pass secrets into the build
  context. In multi-stage builds `ARG`s in earlier stages will not be copied
  to the later stages, or in the metadata of the final stage.
  
  The
  [build metadata](https://concourse-ci.org/implementing-resource-types.html#resource-metadata)
  environment variables provided by Concourse will be expanded in the values
  (the syntax is `$SOME_ENVVAR` or `${SOME_ENVVAR}`).

  Example:

  ```yaml
  build_args:
    DO_THING: true
    HOW_MANY_THINGS: 2
    EMAIL: me@yopmail.com
    CI_BUILD_ID: concourse-$BUILD_ID
  ```

* `build_args_file`: *Optional.* Path to a JSON file containing Docker
  build-time variables.

  Example file contents:

  ```yaml
  { "EMAIL": "me@yopmail.com", "HOW_MANY_THINGS": 1, "DO_THING": false }
  ```

* `cache`: *Optional.* Default `false`. When the `build` parameter is set,
  first pull `image:tag` from the Docker registry (so as to use cached
  intermediate images when building). This will cause the resource to fail
  if it is set to `true` and the image does not exist yet.

* `cache_from`: *Optional.* An array of images to consider as cache, in order to
  reuse build steps from a previous build. The array elements are paths to
  directories generated by a `get` step with `save: true`. This has a similar
  aim of `cache`, but it loads the images from disk instead of pulling them
  from the network, so that Concourse resource caching can be used. It also
  allows more than one image to be specified, which is useful for multi-stage
  Dockerfiles. If you want to cache an image used in a `FROM` step, you should
  put it in `load_bases` instead.

* `cache_tag`: *Optional.* Default `tag`. The specific tag to pull before
  building when `cache` parameter is set. Instead of pulling the same tag
  that's going to be built, this allows picking a different tag like
  `latest` or the previous version. This will cause the resource to fail
  if it is set to a tag that does not exist yet.

* `dockerfile`: *Optional.* The path of the `Dockerfile` in the directory if
  it's not at the root of the directory.

* `docker_buildkit`: *Optional.* This enables a Docker BuildKit build. The value
  should be set to 1 if applicable.

* `import_file`: *Optional.* A path to a file to `docker import` and then push.

* `labels`: *Optional.* A map of labels that will be added to the image.

  Example:

  ```yaml
  labels:
    commit: b4d4823
    version: 1.0.3
  ```

* `labels_file`: *Optional.* Path to a JSON file containing the image labels.

  Example file contents:

  ```json
  { "commit": "b4d4823", "version": "1.0.3" }
  ```

* `load`: *Optional.* The path of a directory containing an image that was
  fetched using this same resource type with `save: true`.

* `load_base`: *Optional.* A path to a directory containing an image to `docker
  load` before running `docker build`. The directory must have `image`,
  `image-id`, `repository`, and `tag` present, i.e. the tree produced by `/in`.

* `load_bases`: *Optional.* Same as `load_base`, but takes an array to load
  multiple images.

* `load_file`: *Optional.* A path to a file to `docker load` and then push.
  Requires `load_repository`.

* `load_repository`: *Optional.* The repository of the image loaded from `load_file`.

* `load_tag`: *Optional.* Default `latest`. The tag of image loaded from `load_file`

* `pull_repository`: *Optional.* **DEPRECATED. Use `get` and `load` instead.** A
  path to a repository to pull down, and then push to this resource.

* `pull_tag`: *Optional.*  **DEPRECATED. Use `get` and `load` instead.** Default
  `latest`. The tag of the repository to pull down via `pull_repository`.

* `tag`: **DEPRECATED - Use `tag_file` instead**
* `tag_file`: *Optional.* The value should be a path to a file containing the name
  of the tag. When not set, the Docker build will be pushed with tag value set by
  `tag` in source configuration.

* `tag_as_latest`: *Optional.*  Default `false`. If true, the pushed image will
  be tagged as `latest` in addition to whatever other tag was specified.

* `tag_prefix`: *Optional.* If specified, the tag read from the file will be
  prepended with this string. This is useful for adding `v` in front of version
  numbers.

* `target_name`: *Optional.*  Specify the name of the target build stage. 
  Only supported for multi-stage Docker builds


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

## Development

### Prerequisites

* golang is *required* - version 1.9.x is tested; earlier versions may also
  work.
* docker is *required* - version 17.06.x is tested; earlier versions may also
  work.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following commands for both `alpine` and `ubuntu` images:

```sh
docker build -t docker-image-resource -f dockerfiles/alpine/Dockerfile .
docker build -t docker-image-resource -f dockerfiles/ubuntu/Dockerfile --build-arg base_image=ubuntu:latest .
```

To use the newly built image, push it to a docker registry that's accessible to
Concourse and configure your pipeline to use it:

```yaml
resource_types:
- name: docker-image-resource
  type: docker-image
  privileged: true
  source:
    repository: example.com:5000/docker-image-resource
    tag: latest

resources:
- name: some-image
  type: docker-image-resource
  ...
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
