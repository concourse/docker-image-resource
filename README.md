# Docker Image Resource

Tracks and builds [Docker](https://docker.io) images.

## Source Configuration

* `repository`: *Required.* The name of the repository, e.g.
`concourse/docker-image-resource`.

* `tag`: *Optional.* The tag to track. Defaults to `latest`.

* `username`: *Optional.* The username to authenticate with when pushing.

* `password`: *Optional.* The password to use when authenticating.

* `email`: *Optional.* The email to use when authenticating.


## Behavior

### `check`: Check for new images.

The current image ID is fetched from the registry for the given tag of the
repository. If it's different from the current version, it is returned.


### `in`: Fetch the image from the registry.

Pulls down the repository from the registry. Note that there's no way to
fetch an image by ID from the Docker regstry, which makes the version
requested irrelevant. Instead, `in` returns the ID of the image that it
ended up fetching as the version.

The following files will be placed in the destination:

* `/image`: The `docker save`d image.
* `/repository`: The name of the repository that was fetched.
* `/tag`: The tag of the repository that was fetched.
* `/image-id`: The fetched image ID.
* `/rootfs.tar`: If `rootfs` is `true`, the contents of the image will be
provided here.

#### Parameters

* `rootfs`: *Optional.* Place a `.tar` file of the image in the destination.


### `out`: Push an image, or build and push a `Dockerfile`.

Push a Docker image to the source's repository and tag. The resulting
version is the image's ID.

#### Parameters

* `push`: *Optional.* Default `true`. Push the image to the Docker index.

* `rootfs`: *Optional.* Default `false`. Place a `.tar` file of the image in the
destination.

* `build`: *Optional.* The path of a directory containing a `Dockerfile` to
build.

* `load_file`: *Optional.* A path to a file to `docker load` and then push.

* `pull_repository`: *Optional.* A path to a repository to pull down, and
then push to this resource.  
