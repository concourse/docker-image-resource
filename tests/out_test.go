package docker_image_resource_test

import (
	"os/exec"

	"encoding/json"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Out", func() {
	BeforeEach(func() {
		os.Setenv("PATH", "/docker-image-resource/tests/fixtures/bin:"+os.Getenv("PATH"))
		os.Setenv("SKIP_PRIVILEGED", "true")
		os.Setenv("LOG_FILE", "/dev/stderr")
	})

	put := func(params map[string]interface{}) *gexec.Session {
		command := exec.Command("/opt/resource/out", "/tmp")

		stdin, err := command.StdinPipe()
		Expect(err).ToNot(HaveOccurred())

		resourceInput, err := json.Marshal(params)
		Expect(err).ToNot(HaveOccurred())
		stdin.Write(resourceInput)
		stdin.Close()

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		session.Wait(10 * time.Second)
		return session
	}

	docker := func(cmd string) string {
		return "DOCKER: " + cmd
	}

	dockerd := func(cmd string) string {
		return "DOCKERD: " + cmd
	}

	It("starts dockerd with --data-root under /scratch", func() {
		session := put(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": "test",
			},
			"params": map[string]interface{}{
				"build": "/docker-image-resource/tests/fixtures/build",
			},
		})

		Expect(session.Err).To(gbytes.Say(dockerd(`.*--data-root /scratch/docker.*`)))
	})

	Context("when configured with a insecure registries", func() {
		It("passes them to dockerd", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository":          "test",
					"insecure_registries": []string{"my-registry.gov", "other-registry.biz"},
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerd(`.*--insecure-registry my-registry\.gov --insecure-registry other-registry\.biz.*`)))
		})
	})

	Context("when configured with a registry mirror", func() {
		It("passes it to dockerd", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository":      "test",
					"registry_mirror": "some-mirror",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerd(`.*--registry-mirror some-mirror.*`)))
		})
	})

	Context("When using ECR", func() {
		It("calls docker pull with the ECR registry", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/ecr",
					"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile",
				},
			})

			Expect(session.Err).To(gbytes.Say(docker("pull 123123.dkr.ecr.us-west-2.amazonaws.com:443/testing")))
		})

		It("calls docker pull for an ECR images in a multi build docker file", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/ecr",
					"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile.multi",
				},
			})

			Expect(session.Err).To(gbytes.Say(docker("pull 123123.dkr.ecr.us-west-2.amazonaws.com:443/testing")))
		})

		Context("When using ECR with dry_run", func() {
			It("successfully builds the image locally", func() {
				session := put(map[string]interface{}{
					"source": map[string]interface{}{
						"repository": "test",
					},
					"params": map[string]interface{}{
						"build":      "/docker-image-resource/tests/fixtures/ecr",
						"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile",
						"dry_run":    "true",
					},
				})
				Expect(session.Out).To(gbytes.Say(docker("Successfully built")))
			})

			It("successfully builds the image multi build docker file locally", func() {
				session := put(map[string]interface{}{
					"source": map[string]interface{}{
						"repository": "test",
					},
					"params": map[string]interface{}{
						"build":      "/docker-image-resource/tests/fixtures/ecr",
						"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile.multi",
						"dry_run":    "true",
					},
				})
				Expect(session.Out).To(gbytes.Say(docker("Successfully built")))
			})
		})
	})
})
