package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var printMetadataPath string

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cmd/print-metadata")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	cmd := exec.Command("make")
	cmd.Dir = ".." // lol

	make, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	<-make.Exited

	Expect(make.ExitCode()).To(Equal(0))

	return []byte("../print-metadata")
}, func(data []byte) {
	printMetadataPath = string(data)
})
