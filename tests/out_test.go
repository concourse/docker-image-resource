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

var _ = Describe("Out", func() {
	BeforeEach(func() {
		os.Setenv("PATH", "/docker-image-resource/tests/fixtures/bin:"+os.Getenv("PATH"))
		os.Setenv("SKIP_PRIVILEGED", "true")
		os.Setenv("LOG_FILE", "/dev/stderr")
	})

	putWithEnv := func(params map[string]interface{}, extraEnv map[string]string) *gexec.Session {
		command := exec.Command("/opt/resource/out", "/tmp")

		// Get current process environment variables
		newEnv := os.Environ()
		if extraEnv != nil {
			// Append each extra environment variable to new process environment
			// variable list
			for name, value := range extraEnv {
				newEnv = append(newEnv, fmt.Sprintf("%s=%s", name, value))
			}
		}

		command.Env = newEnv

		resourceInput, err := json.Marshal(params)
		Expect(err).ToNot(HaveOccurred())

		command.Stdin = bytes.NewBuffer(resourceInput)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		<-session.Exited
		return session
	}

	put := func(params map[string]interface{}) *gexec.Session {
		return putWithEnv(params, nil)
	}

	dockerarg := func(cmd string) string {
		return "DOCKER ARG: " + cmd
	}

	docker := func(cmd string) string {
		return "DOCKER: " + cmd
	}

	dockerd := func(cmd string) string {
		return "DOCKERD: " + cmd
	}

	It("retries starting dockerd if it fails", func() {
		session := putWithEnv(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": "test",
			},
			"params": map[string]interface{}{
				"build": "/docker-image-resource/tests/fixtures/build",
			},
		}, map[string]string{
			"STARTUP_TIMEOUT": "5",
			"FAIL_ONCE": "true",
		})

		Expect(session.Err).To(gbytes.Say("(?s:DOCKERD.*DOCKERD.*)"))
	})

	It("times out retrying dockerd", func() {
		session := putWithEnv(map[string]interface{}{
			"source": map[string]interface{}{
				"repository": "test",
			},
			"params": map[string]interface{}{
				"build": "/docker-image-resource/tests/fixtures/build",
			},
		}, map[string]string{
			"STARTUP_TIMEOUT": "1",
			"FAIL_ONCE": "true",
		})

		Expect(session.Err).To(gbytes.Say(".*Docker failed to start.*"))
	})

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

	Context("when build arguments are provided", func() {
		It("passes the arguments correctly to the docker daemon", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
					"build_args": map[string]string{
						"arg1": "arg with space",
						"arg2": "arg with\nnewline",
						"arg3": "normal",
					},
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg1=arg with space`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg2=arg with\nnewline`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg3=normal`)))
		})
	})

	Context("when secrets are provided", func() {
		It("passes the arguments correctly to the docker daemon", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
					"secrets": map[string]interface{}{
						"secret1": map[string]interface{}{
              "env": "GITHUB_TOKEN",
            },
						"secret2": map[string]interface{}{
              "source": "/a/file/path.txt",
            },
						"secret3": map[string]interface{}{
              "source": "/a/file/path with a space in it.txt",
            },
					},
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--secret`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`id=secret1,env=GITHUB_TOKEN`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--secret`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`id=secret2,source=/a/file/path.txt`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--secret`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`id=secret3,source=/a/file/path with a space in it.txt`)))
		})
	})

	Context("when labels are provided", func() {
		It("passes the labels correctly to the docker daemon", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
					"labels": map[string]string{
						"label1": "foo",
						"label2": "bar\nspam",
						"label3": "eggs eggs eggs",
					},
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--label`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`label1=foo`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--label`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`label2=bar\nspam`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--label`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`label3=eggs eggs eggs`)))
		})
	})

	Context("when build arguments file is provided", func() {
		It("passes the arguments correctly to the docker daemon", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":           "/docker-image-resource/tests/fixtures/build",
					"build_args_file": "/docker-image-resource/tests/fixtures/build_args",
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg1=arg with space`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg2=arg with\nnewline`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg3=normal`)))
		})
	})

	// this is close, but this test is broken. Not sure look for output
	Context("when build arguments file is malformed", func() {
		It("provides a useful error message", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":           "/docker-image-resource/tests/fixtures/build",
					"build_args_file": "/docker-image-resource/tests/fixtures/build_args_malformed",
				},
			})

			Expect(session.Err).To(gbytes.Say(`Failed to parse build_args_file \(/docker-image-resource/tests/fixtures/build_args_malformed\)`))
		})
	})

	Context("when build arguments contain envvars", func() {
		It("expands envvars and passes the arguments correctly to the docker daemon", func() {
			session := putWithEnv(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
					"build_args": map[string]string{
						"arg01": "no envvars",
						"arg02": "this is the:\n$BUILD_ID",
						"arg03": "this is the:\n$BUILD_NAME",
						"arg04": "this is the:\n$BUILD_JOB_NAME",
						"arg05": "this is the:\n$BUILD_PIPELINE_NAME",
						"arg06": "this is the:\n$BUILD_TEAM_NAME",
						"arg07": "this is the:\n$ATC_EXTERNAL_URL",
						"arg08": "this syntax also works:\n${ATC_EXTERNAL_URL}",
						"arg09": "multiple envvars in one arg also works:\n$BUILD_ID\n$BUILD_NAME",
						"arg10": "$BUILD_ID at the beginning of the arg",
						"arg11": "at the end of the arg is the $BUILD_ID",
					},
				},
			}, map[string]string{
				"BUILD_ID":		"value of the:\nBUILD_ID envvar",
				"BUILD_NAME":		"value of the:\nBUILD_NAME envvar",
				"BUILD_JOB_NAME":	"value of the:\nBUILD_JOB_NAME envvar",
				"BUILD_PIPELINE_NAME":	"value of the:\nBUILD_PIPELINE_NAME envvar",
				"BUILD_TEAM_NAME":	"value of the:\nBUILD_TEAM_NAME envvar",
				"ATC_EXTERNAL_URL":	"value of the:\nATC_EXTERNAL_URL envvar",
			})


			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg01=no envvars`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg02=this is the:\nvalue of the:\nBUILD_ID envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg03=this is the:\nvalue of the:\nBUILD_NAME envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg04=this is the:\nvalue of the:\nBUILD_JOB_NAME envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg05=this is the:\nvalue of the:\nBUILD_PIPELINE_NAME envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg06=this is the:\nvalue of the:\nBUILD_TEAM_NAME envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg07=this is the:\nvalue of the:\nATC_EXTERNAL_URL envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg08=this syntax also works:\nvalue of the:\nATC_EXTERNAL_URL envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg09=multiple envvars in one arg also works:\nvalue of the:\nBUILD_ID envvar\nvalue of the:\nBUILD_NAME envvar`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg10=value of the:\nBUILD_ID envvar at the beginning of the arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg11=at the end of the arg is the value of the:\nBUILD_ID envvar`)))
		})
	})

	Context("when configured with limited up and download", func() {
		It("passes them to dockerd", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository":               "test",
					"max_concurrent_downloads": 7,
					"max_concurrent_uploads":   1,
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerd(`.* --max-concurrent-downloads=7 --max-concurrent-uploads=1.*`)))
		})
	})

	Context("when configured with additional private registries", func() {
		It("passes them to docker login", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
					"additional_private_registries": []interface{}{
						map[string]string{
							"registry": "example.com/my-private-docker-registry",
							"username": "my-username",
							"password": "my-secret",
						},
						map[string]string{
							"registry": "example.com/another-private-docker-registry",
							"username": "another-username",
							"password": "another-secret",
						},
					},
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`login`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`-u`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`my-username`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--password-stdin`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`example.com/my-private-docker-registry`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`login`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`-u`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`another-username`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--password-stdin`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`example.com/another-private-docker-registry`)))
		})
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

	Context("when configured with a target build stage", func() {
		It("passes it to dockerd", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"target_name": "test",
					"build":       "/docker-image-resource/tests/fixtures/build",
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--target`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`test`)))
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

		It("calls docker pull for an ECR image in a multi build docker file", func() {
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

		It("calls docker pull for all ECR images in a multi build docker file", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/ecr",
					"dockerfile": "/docker-image-resource/tests/fixtures/ecr/Dockerfile.multi-ecr",
				},
			})

			Expect(session.Err).To(gbytes.Say(docker("pull 123123.dkr.ecr.us-west-2.amazonaws.com:443/testing")))
			Expect(session.Err).To(gbytes.Say(docker("pull 123123.dkr.ecr.us-west-2.amazonaws.com:443/testing2")))
		})
	})

	Context("When all proxy settings are provided with build args", func() {
		It("passes the arguments correctly to the docker daemon", func() {
			session := putWithEnv(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
					"build_args": map[string]string{
						"arg1": "arg with space",
						"arg2": "arg with\nnewline",
						"arg3": "normal",
					},
				},
			}, map[string]string{
				"no_proxy":    "10.1.1.1",
				"http_proxy":  "http://admin:verysecret@my-proxy.com:8080",
				"https_proxy": "http://another.proxy.net",
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`http_proxy=http://admin:verysecret@my-proxy.com:8080`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`https_proxy=http://another.proxy.net`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`no_proxy=10.1.1.1`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg1=arg with space`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg2=arg with\nnewline`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`arg3=normal`)))
		})
	})

	Context("When passing tag ", func() {
		It("should pull tag from file", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
					"tag":   "/docker-image-resource/tests/fixtures/tag",
				},
			},
			)
			Expect(session.Err).To(gbytes.Say(docker(`push test:foo`)))
		})
	})

	Context("When passing tag_file", func() {
		It("should pull tag from file", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":    "/docker-image-resource/tests/fixtures/build",
					"tag_file": "/docker-image-resource/tests/fixtures/tag",
				},
			},
			)
			Expect(session.Err).To(gbytes.Say(docker(`push test:foo`)))
		})
	})

	Context("When passing tag and tag_file", func() {
		It("should pull tag from file (prefer tag_file param)", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":    "/docker-image-resource/tests/fixtures/build",
					"tag":      "/docker-image-resource/tests/fixtures/doesnotexist",
					"tag_file": "/docker-image-resource/tests/fixtures/tag",
				},
			},
			)
			Expect(session.Err).To(gbytes.Say(docker(`push test:foo`)))
		})
	})

	Context("When passing additional_tags ", func() {
		It("should push add the additional_tags", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":           "/docker-image-resource/tests/fixtures/build",
					"additional_tags": "/docker-image-resource/tests/fixtures/tags",
				},
			},
			)
			Expect(session.Err).To(gbytes.Say(docker(`push test:latest`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag test:latest test:a`)))
			Expect(session.Err).To(gbytes.Say(docker(`push test:a`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag test:latest test:b`)))
			Expect(session.Err).To(gbytes.Say(docker(`push test:b`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag test:latest test:c`)))
			Expect(session.Err).To(gbytes.Say(docker(`push test:c`)))
		})
	})

	Context("When only http_proxy setting is provided, with no build arguments", func() {
		It("passes the arguments correctly to the docker daemon", func() {
			session := putWithEnv(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build": "/docker-image-resource/tests/fixtures/build",
				},
			}, map[string]string{
				"http_proxy": "http://admin:verysecret@my-proxy.com:8080",
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--build-arg`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`http_proxy=http://admin:verysecret@my-proxy.com:8080`)))
		})
	})

	Context("when load_bases are specified", func() {
		BeforeEach(func() {
			os.Mkdir("/tmp/expected_base_1", os.ModeDir)
			// this image should really be an actual tarball, but the test passes with text. :shrug:
			os.WriteFile("/tmp/expected_base_1/image", []byte("some-image-1"), os.ModePerm)
			os.WriteFile("/tmp/expected_base_1/repository", []byte("some-repository-1"), os.ModePerm)
			os.WriteFile("/tmp/expected_base_1/image-id", []byte("some-image-id-1"), os.ModePerm)
			os.WriteFile("/tmp/expected_base_1/tag", []byte("some-tag-1"), os.ModePerm)

			os.Mkdir("/tmp/expected_base_2", os.ModeDir)
			os.WriteFile("/tmp/expected_base_2/image", []byte("some-image-2"), os.ModePerm)
			os.WriteFile("/tmp/expected_base_2/repository", []byte("some-repository-2"), os.ModePerm)
			os.WriteFile("/tmp/expected_base_2/image-id", []byte("some-image-id-2"), os.ModePerm)
			os.WriteFile("/tmp/expected_base_2/tag", []byte("some-tag-2"), os.ModePerm)

			os.Mkdir("/tmp/unexpected_base", os.ModeDir)
			os.WriteFile("/tmp/unexpected_base/image", []byte("some-image-3"), os.ModePerm)
			os.WriteFile("/tmp/unexpected_base/repository", []byte("some-repository-3"), os.ModePerm)
			os.WriteFile("/tmp/unexpected_base/image-id", []byte("some-image-id-3"), os.ModePerm)
			os.WriteFile("/tmp/unexpected_base/tag", []byte("some-tag-3"), os.ModePerm)
		})

		AfterEach(func() {
			os.RemoveAll("/tmp/expected_base_1")
			os.RemoveAll("/tmp/expected_base_2")
			os.RemoveAll("/tmp/unexpected_base")
		})

		It("passes the arguments correctly to the docker daemon", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/build",
					"load_bases": []string{"expected_base_1", "expected_base_2"},
				},
			})

			Expect(session.Err).To(gbytes.Say(docker(`load -i expected_base_1/image`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag some-image-id-1 some-repository-1:some-tag-1`)))

			Expect(session.Err).To(gbytes.Say(docker(`load -i expected_base_2/image`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag some-image-id-2 some-repository-2:some-tag-2`)))

			Expect(session.Err).NotTo(gbytes.Say(docker(`load -i unexpected_base/image`)))
			Expect(session.Err).NotTo(gbytes.Say(docker(`tag some-image-id-3 some-repository-3:some-tag-3`)))
		})
	})

	Context("when cache is specified are specified", func() {
		It("adds argument to cache_from", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/build",
					"cache":      "true",
				},
			})
			Expect(session.Err).To(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`test:latest`)))
		})

		It("does not add cache_from if pull fails", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "broken-repo",
				},
				"params": map[string]interface{}{
					"build":     "/docker-image-resource/tests/fixtures/build",
					"cache":      "true",
				},
			})

			Expect(session.Err).ToNot(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`broken-repo:latest`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`build`)))

		})
	});

	Context("when cache_from images are specified", func() {
		BeforeEach(func() {
			os.Mkdir("/tmp/cache_from_1", os.ModeDir)
			// this image should really be an actual tarball, but the test passes with text. :shrug:
			os.WriteFile("/tmp/cache_from_1/image", []byte("some-image-1"), os.ModePerm)
			os.WriteFile("/tmp/cache_from_1/repository", []byte("some-repository-1"), os.ModePerm)
			os.WriteFile("/tmp/cache_from_1/image-id", []byte("some-image-id-1"), os.ModePerm)
			os.WriteFile("/tmp/cache_from_1/tag", []byte("some-tag-1"), os.ModePerm)

			os.Mkdir("/tmp/cache_from_2", os.ModeDir)
			os.WriteFile("/tmp/cache_from_2/image", []byte("some-image-2"), os.ModePerm)
			os.WriteFile("/tmp/cache_from_2/repository", []byte("some-repository-2"), os.ModePerm)
			os.WriteFile("/tmp/cache_from_2/image-id", []byte("some-image-id-2"), os.ModePerm)
			os.WriteFile("/tmp/cache_from_2/tag", []byte("some-tag-2"), os.ModePerm)
		})

		AfterEach(func() {
			os.RemoveAll("/tmp/cache_from_1")
			os.RemoveAll("/tmp/cache_from_2")
		})

		It("calls docker load to load the cache_from images", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				}, 
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/build",
					"cache_from": []string{"cache_from_1", "cache_from_2"},
				},
			})
			Expect(session.Err).To(gbytes.Say(docker(`load -i cache_from_1/image`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag some-image-id-1 some-repository-1:some-tag-1`)))

			Expect(session.Err).To(gbytes.Say(docker(`load -i cache_from_2/image`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag some-image-id-2 some-repository-2:some-tag-2`)))
		})

		It("loads both cache_from and load_bases images", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/build",
					"load_bases": []string{"cache_from_1"},
					"cache_from": []string{"cache_from_2"},
				},
			})
			Expect(session.Err).To(gbytes.Say(docker(`load -i cache_from_1/image`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag some-image-id-1 some-repository-1:some-tag-1`)))

			Expect(session.Err).To(gbytes.Say(docker(`load -i cache_from_2/image`)))
			Expect(session.Err).To(gbytes.Say(docker(`tag some-image-id-2 some-repository-2:some-tag-2`)))
		})

		It("passes the arguments correctly to the docker build command", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/build",
					"cache_from": []string{"cache_from_1", "cache_from_2"},
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`some-repository-1:some-tag-1`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`some-repository-2:some-tag-2`)))
		})

		It("does not remove the arguments generated by cache:true", func() {
			session := put(map[string]interface{}{
				"source": map[string]interface{}{
					"repository": "test",
				},
				"params": map[string]interface{}{
					"build":      "/docker-image-resource/tests/fixtures/build",
					"cache":      "true",
					"cache_from": []string{"cache_from_1", "cache_from_2"},
				},
			})

			Expect(session.Err).To(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`test:latest`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`some-repository-1:some-tag-1`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`--cache-from`)))
			Expect(session.Err).To(gbytes.Say(dockerarg(`some-repository-2:some-tag-2`)))
		})
	})
})
