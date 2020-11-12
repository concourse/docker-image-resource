This fork adds the ability to build a Docker config.json into the image.
We use this to default to using the ecr-login credential helper for all our AWS account's registries.

To build this project first retrieve the config.json from 1Password, place it in the project root, and then build the docker image.
