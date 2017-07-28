package docker_image_resource_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"encoding/json"
	"os"
	"github.com/onsi/gomega/gbytes"
	"time"
)

var source = map[string]interface{} {
	"repository": "testreg:5000/test",
	"insecure_registries": [1]string{"testreg:5000"},
}

var _ = Describe("Out", func() {
	BeforeEach(func() {
		os.Setenv("PATH", "/docker-image-resource/tests/fixtures/bin:" + os.Getenv("PATH"))
	})

	put := func(params map[string]interface{}) *gexec.Session {
		command := exec.Command("/docker-image-resource/assets/out", "/tmp")

		stdin, err := command.StdinPipe()
		Expect(err).ToNot(HaveOccurred())

		resourceInput, err := json.Marshal(params)
		Expect(err).ToNot(HaveOccurred())
		stdin.Write(resourceInput)
		stdin.Close()

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		session.Wait(30 * time.Second)
		return session
	}

	docker := func(cmd string) string {
		return "DOCKER: " + cmd
	}

	Context("When using ECR", func() {
		It("calls docker pull with the ECR registry", func() {
			session := put(map[string]interface{} {
				"source": map[string]interface{} {
					"repository": "test",
				},
				"params": map[string]interface{} {
					"build" : "/docker-image-resource/tests/fixtures/ecr",
					"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile",
				},
			})

			Expect(session.Err).To(gbytes.Say(docker("pull 123123.dkr.ecr.us-west-2.amazonaws.com:443/testing")))
		})

		It("calls docker pull for an ECR images in a multi build docker file", func() {
			session := put(map[string]interface{} {
				"source": map[string]interface{} {
					"repository": "test",
				},
				"params": map[string]interface{} {
					"build" : "/docker-image-resource/tests/fixtures/ecr",
					"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile.multi",
				},
			})

			Expect(session.Err).To(gbytes.Say(docker("pull 123123.dkr.ecr.us-west-2.amazonaws.com:443/testing")))
		})
	})

	Context("When tagging as latest", func() {
		var fixdir = "/docker-image-resource/tests/fixtures/basic"
		It("doesn't tag the image with :latest by default", func() {
			session := put(map[string]interface{} {
				"source": source,
				"params": map[string]interface{} {
					"build": fixdir,
					"tag": fixdir + "/tag",
					"dockerfile": fixdir + "/Dockerfile",
				},
			})

			Expect(session.Err).To(gbytes.Say("Successfully tagged testreg:5000/test:mytag"))
			Expect(session.Err).ToNot(gbytes.Say("testreg:5000/test:mytag tagged as latest"))
		})
		It("tags the image with :latest", func() {
			session := put(map[string]interface{} {
				"source": source,
				"params": map[string]interface{} {
					"build": fixdir,
					"tag": fixdir + "/tag",
					"dockerfile": fixdir + "/Dockerfile",
					"tag_as_latest": true,
				},
			})

			Expect(session.Err).To(gbytes.Say("Successfully tagged testreg:5000/test:mytag"))
			Expect(session.Err).To(gbytes.Say("testreg:5000/test:mytag tagged as latest"))
		})
	})
})
