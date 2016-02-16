package passwd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPasswd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Passwd Suite")
}
