package imageplugin_test

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"

	"code.cloudfoundry.org/garden-shed/rootfs_provider"
	"code.cloudfoundry.org/guardian/imageplugin"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ExternalImageManager", func() {
	var (
		fakeCommandRunner    *fake_command_runner.FakeCommandRunner
		logger               *lagertest.TestLogger
		externalImageManager *imageplugin.ExternalImageManager
		imageSource          *url.URL
		idMappings           []specs.IDMapping
		defaultRootFS        *url.URL
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("external-image-manager")
		fakeCommandRunner = fake_command_runner.New()

		idMappings = []specs.IDMapping{
			specs.IDMapping{
				ContainerID: 0,
				HostID:      100,
				Size:        1,
			},
			specs.IDMapping{
				ContainerID: 1,
				HostID:      1,
				Size:        99,
			},
		}

		var err error
		defaultRootFS, err = url.Parse("/default/rootfs")
		Expect(err).ToNot(HaveOccurred())
		externalImageManager = imageplugin.New("/external-image-manager-bin", fakeCommandRunner, defaultRootFS, idMappings)

		imageSource, err = url.Parse("/hello/image")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Create", func() {
		var (
			returnedRootFS string
			testQuotaSize  int64
			err            error
			namespaced     bool
		)

		BeforeEach(func() {
			testQuotaSize = 0
			namespaced = false
		})

		JustBeforeEach(func() {
			returnedRootFS, _, err = externalImageManager.Create(logger, "hello", rootfs_provider.Spec{
				QuotaSize:  testQuotaSize,
				RootFS:     imageSource,
				Namespaced: namespaced,
			})
		})

		It("uses the correct external-image-manager binary", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
			imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

			Expect(imageManagerCmd.Path).To(Equal("/external-image-manager-bin"))
		})

		Describe("external-image-manager parameters", func() {
			It("uses the correct external-image-manager create command", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
				imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

				Expect(imageManagerCmd.Args[1]).To(Equal("create"))
			})

			It("sets the correct image input to external-image-manager", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
				imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

				Expect(imageManagerCmd.Args[len(imageManagerCmd.Args)-2]).To(Equal("/hello/image"))
			})

			It("sets the correct id to external-image-manager", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
				imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

				Expect(imageManagerCmd.Args[len(imageManagerCmd.Args)-1]).To(Equal("hello"))
			})

			Context("when namespaced is true", func() {
				BeforeEach(func() {
					namespaced = true
				})

				It("passes the correct uid and gid mappings to the external-image-manager", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
					imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

					firstMap := fmt.Sprintf("%d:%d:%d", idMappings[0].ContainerID, idMappings[0].HostID, idMappings[0].Size)
					secondMap := fmt.Sprintf("%d:%d:%d", idMappings[1].ContainerID, idMappings[1].HostID, idMappings[1].Size)

					Expect(imageManagerCmd.Args[2:10]).To(Equal([]string{
						"--uid-mapping", firstMap,
						"--gid-mapping", firstMap,
						"--uid-mapping", secondMap,
						"--gid-mapping", secondMap,
					}))
				})

				It("runs the external-image-manager as the container root user", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
					imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

					Expect(imageManagerCmd.SysProcAttr.Credential.Uid).To(Equal(idMappings[0].HostID))
					Expect(imageManagerCmd.SysProcAttr.Credential.Gid).To(Equal(idMappings[0].HostID))
				})
			})

			Context("when namespaced is false", func() {
				It("does not pass any uid and gid mappings to the external-image-manager", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
					imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

					Expect(imageManagerCmd.Args).NotTo(ContainElement("--uid-mapping"))
					Expect(imageManagerCmd.Args).NotTo(ContainElement("--gid-mapping"))
				})
			})

			Context("when a disk quota is provided in the spec", func() {
				BeforeEach(func() {
					testQuotaSize = 1024
				})

				It("passes the quota to the external-image-manager", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
					imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

					Expect(imageManagerCmd.Args[2]).To(Equal("--disk-limit-size-bytes"))
					Expect(imageManagerCmd.Args[3]).To(Equal(strconv.FormatInt(testQuotaSize, 10)))
				})
			})
		})

		Context("when the external-image-manager binary prints to stdout/stderr", func() {
			BeforeEach(func() {
				fakeCommandRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/external-image-manager-bin",
				}, func(cmd *exec.Cmd) error {
					cmd.Stdout.Write([]byte("/this-is/your"))
					cmd.Stderr.Write([]byte("/this-is-not/your-rootfs"))
					return nil
				})
			})

			It("returns stdout as the rootfs location", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(returnedRootFS).To(Equal("/this-is/your/rootfs"))
			})
		})

		Context("when the external-image-manager binary prints a newline ath the end of its output", func() {
			BeforeEach(func() {
				fakeCommandRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/external-image-manager-bin",
				}, func(cmd *exec.Cmd) error {
					cmd.Stdout.Write([]byte("/this-is/your\n"))
					return nil
				})
			})

			It("returns the rootfs without the new line character", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(returnedRootFS).To(Equal("/this-is/your/rootfs"))
			})
		})

		Context("when the command fails", func() {
			BeforeEach(func() {
				fakeCommandRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/external-image-manager-bin",
				}, func(cmd *exec.Cmd) error {
					cmd.Stderr.Write([]byte("btrfs doesn't like you"))

					return errors.New("external-image-manager failure")
				})
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("external image manager create failed")))
				Expect(err).To(MatchError(ContainSubstring("external-image-manager failure")))
			})

			It("returns the external-image-manager error output in the error", func() {
				Expect(logger).To(gbytes.Say("btrfs doesn't like you"))
			})
		})

		Context("when a RootFS is not provided in the rootfs_provider.Spec", func() {
			BeforeEach(func() {
				imageSource = &url.URL{}
			})

			It("passes the default rootfs to the external-image-manager", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
				imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

				Expect(imageManagerCmd.Args[len(imageManagerCmd.Args)-2]).To(Equal(defaultRootFS.String()))
			})
		})
	})

	Describe("Destroy", func() {
		var err error

		JustBeforeEach(func() {
			err = externalImageManager.Destroy(logger, "hello")
		})

		It("uses the correct external-image-manager binary", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
			imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

			Expect(imageManagerCmd.Path).To(Equal("/external-image-manager-bin"))
		})

		Describe("external-image-manager parameters", func() {
			It("uses the correct external-image-manager delete command", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
				imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

				Expect(imageManagerCmd.Args[1]).To(Equal("delete"))
			})

			It("sets the correct id to external-image-manager", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCommandRunner.ExecutedCommands())).To(Equal(1))
				imageManagerCmd := fakeCommandRunner.ExecutedCommands()[0]

				Expect(imageManagerCmd.Args[len(imageManagerCmd.Args)-1]).To(Equal("hello"))
			})
		})

		Context("when the command fails", func() {
			BeforeEach(func() {
				fakeCommandRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/external-image-manager-bin",
				}, func(cmd *exec.Cmd) error {
					cmd.Stderr.Write([]byte("btrfs doesn't like you"))

					return errors.New("external-image-manager failure")
				})
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("external image manager destroy failed")))
				Expect(err).To(MatchError(ContainSubstring("external-image-manager failure")))
			})

			It("returns the external-image-manager error output in the error", func() {
				Expect(logger).To(gbytes.Say("btrfs doesn't like you"))
			})
		})
	})
})
