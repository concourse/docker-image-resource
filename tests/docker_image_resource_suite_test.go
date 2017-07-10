package docker_image_resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDockerImageResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DockerImageResource Suite")
}
