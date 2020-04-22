package docker_image_resource_test

import (
	"bytes"
	"fmt"
	"os/exec"

	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Check", func() {
	BeforeEach(func() {
		os.Setenv("PATH", "/docker-image-resource/tests/fixtures/bin:"+os.Getenv("PATH"))
		os.Setenv("SKIP_PRIVILEGED", "true")
		os.Setenv("LOG_FILE", "/dev/stderr")
	})

	check := func(params map[string]interface{}) *gexec.Session {
		command := exec.Command("/opt/resource/check", "/tmp")

		resourceInput, err := json.Marshal(params)
		Expect(err).ToNot(HaveOccurred())

		command.Stdin = bytes.NewBuffer(resourceInput)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		<-session.Exited
		return session
	}

	It("errors when image is unknown", func() {
		repository := "kjlasdfaklklj12"
		tag := "latest"
		session := check(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": repository,
			},
		})

		expectedStringInError := fmt.Sprintf("%s:%s", repository, tag)
		Expect(session.Err).To(gbytes.Say(expectedStringInError))
	})

	It("errors when image:tag is unknown", func() {
		repository := "kjlasdfaklklj12"
		tag := "aklsdf123"
		session := check(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": repository,
				"tag":        tag,
			},
		})

		expectedStringInError := fmt.Sprintf("%s:%s", repository, tag)
		Expect(session.Err).To(gbytes.Say(expectedStringInError))
	})

	It("prints out the digest for a existent image", func() {
		session := check(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": "alpine",
			},
		})

		Expect(session.Out).To(gbytes.Say(`{"digest":`))
	})

	It("prints out the digest for a existent image and quoted numeric tag", func() {
		session := check(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": "alpine",
				"tag":        "3.7",
			},
		})

		Expect(session.Out).To(gbytes.Say(`{"digest":`))
	})

	It("prints out the digest for a existent image and numeric tag", func() {
		session := check(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": "alpine",
				"tag":        3.7,
			},
		})

		Expect(session.Out).To(gbytes.Say(`{"digest":`))
	})
})
