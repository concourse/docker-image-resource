package docker_image_resource_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"

	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
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

	Context("when a registry mirror is configured", func() {
		var (
			registry         *ghttp.Server
			session          *gexec.Session
			latestFakeDigest string = "sha256:c4c25c2cd70e3071f08cf124c4b5c656c061dd38247d166d97098d58eeea8aa6"
		)

		BeforeEach(func() {
			registry = ghttp.NewServer()

			BeforeEach(func() {
				registry.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/"),
						ghttp.RespondWith(http.StatusOK, "fake mirror"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("HEAD", "/v2/some/fake-image/manifests/latest"),
						ghttp.VerifyHeaderKV("Accept",
							"application/vnd.docker.distribution.manifest.v2+json",
							"application/vnd.oci.image.index.v1+json",
							"application/json",
						),
						ghttp.RespondWith(http.StatusOK, `{"fake":"manifest"}`, http.Header{
							"Docker-Content-Digest": {latestFakeDigest},
						}),
					),
				)
			})

			AfterEach(func() {
				registry.Close()
			})

			Context("when the repository contains no registry hostname", func() {
				BeforeEach(func() {
					session = check(map[string]interface{}{
						"source": map[string]interface{}{
							"repository":      "some/fake-image",
							"registry_mirror": registry.URL(),
						},
					})
				})

				It("fetches the image data from the mirror and prints out the digest and tag", func() {
					Expect(session.Out).To(gbytes.Say(fmt.Sprintf(`[{"digest":"%s"}]`, latestFakeDigest)))
				})

				It("does not error", func() {
					Expect(session.Err).To(Equal(""))
				})
			})

			Context("when the repository contains a registry hostname different from the mirror", func() {
				BeforeEach(func() {
					session = check(map[string]interface{}{
						"source": map[string]interface{}{
							"repository":      registry.URL() + "/some/fake-image",
							"registry_mirror": "https://thisregistrydoesnotexist.nothing",
						},
					})
				})

				It("fetches the image data from the registry cited in the repository rather than the mirror and prints out the digest and tag", func() {
					Expect(session.Out).To(gbytes.Say(fmt.Sprintf(`[{"digest":"%s"}]`, latestFakeDigest)))
				})

				It("does not error", func() {
					Expect(session.Err).To(Equal(""))
				})
			})
		})
	})
})
