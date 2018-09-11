package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

type imageMetadata struct {
	User string   `json:"user"`
	Env  []string `json:"env"`
}

var _ = Describe("print-metadata", func() {
	var (
		cmd      *exec.Cmd
		userFile *os.File

		metadata imageMetadata
	)

	JustBeforeEach(func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		metadata = imageMetadata{}
		err = json.Unmarshal(session.Out.Contents(), &metadata)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when user file exists", func() {
		BeforeEach(func() {
			var err error
			userFile, err = ioutil.TempFile("", "print-metadata-test")
			Expect(err).NotTo(HaveOccurred())

			cmd = exec.Command(printMetadataPath, "-userFile", userFile.Name())
		})

		AfterEach(func() {
			userFile.Close()
			os.Remove(userFile.Name())
		})

		It("writes metadata with no user", func() {
			Expect(metadata.User).To(BeEmpty())
		})

		Context("when password file contains current user", func() {
			BeforeEach(func() {
				currentUsedID := syscall.Getuid()

				_, err := userFile.WriteString(fmt.Sprintf(
					"%s:*:%d:%d:System Administrator:/var/%s:/bin/sh\n",
					"some-user",
					currentUsedID,
					currentUsedID,
					"some-user",
				))
				Expect(err).NotTo(HaveOccurred())
				userFile.Sync()
			})

			It("sets current user in metadata", func() {
				Expect(metadata.User).To(Equal("some-user"))
			})
		})
	})

	Context("when password file does not exist", func() {
		BeforeEach(func() {
			cmd = exec.Command(printMetadataPath, "-userFile", "non-existent-file")
		})

		It("writes metadata with no user", func() {
			Expect(metadata.User).To(BeEmpty())
		})
	})

	Describe("environment variables", func() {
		BeforeEach(func() {
			if runtime.GOOS == "darwin" && syscall.Getuid() != 0 {
				Skip("OS X doesn't use /etc/passwd for multi-user mode (you need to run the tests as root)")
			}
			cmd = exec.Command(printMetadataPath)
		})

		Context("when it is running in an environment with environment variables", func() {
			BeforeEach(func() {
				cmd.Env = []string{
					"SOME=foo",
					"AMAZING=bar",
					"ENV=baz",
				}
			})

			It("outputs them on stdout", func() {
				Expect(metadata.Env).To(ConsistOf([]string{
					"SOME=foo",
					"AMAZING=bar",
					"ENV=baz",
				}))
			})
		})

		Context("when it is running in an environment with environment variables in the blacklist", func() {
			BeforeEach(func() {
				cmd.Env = []string{
					"SOME=foo",
					"HOSTNAME=bar",
					"ENV=baz",
				}
			})

			It("outputs everything but them on stdout", func() {
				Expect(metadata.Env).To(ConsistOf([]string{
					"SOME=foo",
					"ENV=baz",
				}))
			})
		})
	})
})
