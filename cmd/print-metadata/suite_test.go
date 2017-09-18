package main_test

import (
	"os"
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
	var path string
	if _, err := os.Stat("/opt/resource/print-metadata"); err == nil {
		path = "/opt/resource/print-metadata"
	} else {
		path, err = gexec.Build("github.com/concourse/docker-image-resource/cmd/print-metadata")
		Expect(err).NotTo(HaveOccurred())
	}

	return []byte(path)
}, func(data []byte) {
	printMetadataPath = string(data)
})
