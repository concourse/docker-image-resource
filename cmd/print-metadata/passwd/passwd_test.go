package passwd_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/docker-image-resource/cmd/print-metadata/passwd"
)

var _ = Describe("Passwd", func() {
	var (
		etcPasswdDir      string
		etcPasswdPath     string
		etcPasswdContents string
		etcPasswdUsers    []passwd.User
	)

	BeforeEach(func() {
		etcPasswdContents = ""
		etcPasswdUsers = []passwd.User{}
	})

	JustBeforeEach(func() {
		path, err := ioutil.TempDir("", "passwd")
		Expect(err).ToNot(HaveOccurred())

		etcPasswdDir = path
		etcPasswdPath = filepath.Join(etcPasswdDir, "passwd")

		for _, user := range etcPasswdUsers {
			etcPasswdContents += fmt.Sprintf("%s:*:%d:1:User Name:/dev/null:/usr/bin/false\n", user.Username, user.ID)
		}

		err = ioutil.WriteFile(etcPasswdPath, []byte(etcPasswdContents), 0600)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(etcPasswdDir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("getting a list of users", func() {
		Context("when there is a single user in the passwd file", func() {
			BeforeEach(func() {
				etcPasswdUsers = []passwd.User{
					{
						ID:       1,
						Username: "username",
					},
				}
			})

			It("can read the specified passwd file", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).ToNot(BeNil())
			})

			It("finds one user", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(ConsistOf(etcPasswdUsers))
			})
		})

		Context("when there is a different single user in the passwd file", func() {
			BeforeEach(func() {
				etcPasswdUsers = []passwd.User{
					{
						ID:       2,
						Username: "username2",
					},
				}
			})

			It("finds the different user", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(ConsistOf(etcPasswdUsers))
			})
		})

		Context("when there are multiple users in the passwd file", func() {
			BeforeEach(func() {
				etcPasswdUsers = []passwd.User{
					{
						ID:       1,
						Username: "username",
					},
					{
						ID:       2,
						Username: "username2",
					},
				}
			})

			It("finds two users", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(ConsistOf(etcPasswdUsers))
			})
		})

		Context("when the file contains comments", func() {
			BeforeEach(func() {
				etcPasswdContents = `# this is a comment
commentuser:*:1:1:User Name:/dev/null:/usr/bin/false\n
			`
			})

			It("finds the user", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(ConsistOf(passwd.User{
					ID:       1,
					Username: "commentuser",
				}))
			})
		})

		Context("when the file contains comments that have whitespace before them", func() {
			BeforeEach(func() {
				etcPasswdContents = `  # this is a comment
commentuser:*:1:1:User Name:/dev/null:/usr/bin/false\n
			`
			})

			It("finds the user", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(ConsistOf(passwd.User{
					ID:       1,
					Username: "commentuser",
				}))
			})
		})

		Context("when the file contains malformed user lines", func() {
			BeforeEach(func() {
				etcPasswdContents = `

			commentuser:*:1:::::1:User Name:/dev/null:/usr/bin/false\n
			`
			})

			It("returns an error", func() {
				_, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).To(MatchError("malformed user on line 3"))
			})
		})

		Context("when the file does not exist", func() {
			It("returns an error", func() {
				_, err := passwd.ReadUsers("/this/does/not/exist")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the file contains a malformed user ID", func() {
			BeforeEach(func() {
				etcPasswdContents = `

			commentuser:*:hello:1:User Name:/dev/null:/usr/bin/false\n
			`
			})

			It("returns an error", func() {
				_, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).To(MatchError("malformed user ID on line 3: hello"))
			})
		})

		Context("when the file does not exist", func() {
			It("returns an error", func() {
				_, err := passwd.ReadUsers("/this/does/not/exist")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("getting a username from a user id", func() {
		Context("when the user exists", func() {
			BeforeEach(func() {
				etcPasswdUsers = []passwd.User{
					{
						ID:       1,
						Username: "username",
					},
					{
						ID:       2,
						Username: "username2",
					},
				}
			})

			It("lets someone find the username of a particular ID", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())

				name, found := users.NameForID(2)
				Expect(found).To(BeTrue())
				Expect(name).To(Equal("username2"))
			})
		})

		Context("when a different user exists", func() {
			BeforeEach(func() {
				etcPasswdUsers = []passwd.User{
					{
						ID:       1,
						Username: "username",
					},
					{
						ID:       2,
						Username: "username2",
					},
				}
			})

			It("lets someone find the username of a particular ID", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())

				name, found := users.NameForID(1)
				Expect(found).To(BeTrue())
				Expect(name).To(Equal("username"))
			})
		})

		Context("when the user doesn't exist", func() {
			BeforeEach(func() {
				etcPasswdUsers = []passwd.User{
					{
						ID:       1,
						Username: "username",
					},
					{
						ID:       2,
						Username: "username2",
					},
				}
			})

			It("lets someone find the username of a particular ID", func() {
				users, err := passwd.ReadUsers(etcPasswdPath)
				Expect(err).ToNot(HaveOccurred())

				_, found := users.NameForID(3)
				Expect(found).To(BeFalse())
			})
		})
	})
})
