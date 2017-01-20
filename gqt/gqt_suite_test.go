package gqt_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/guardian/gqt/runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"encoding/json"
	"testing"
)

var defaultRuntime = map[string]string{
	"linux": "runc",
}

var ginkgoIO = garden.ProcessIO{Stdout: GinkgoWriter, Stderr: GinkgoWriter}

var (
	ociRuntimeBin      string
	gardenBin          string
	initBin            string
	nstarBin           string
	dadooBin           string
	testImagePluginBin string
	inspectorGardenBin string
	testNetPluginBin   string
	tarBin             string
	containerdShimBin  string
)

func TestGqt(t *testing.T) {
	RegisterFailHandler(Fail)

	SynchronizedBeforeSuite(func() []byte {
		var err error
		bins := make(map[string]string)

		bins["oci_runtime_path"] = os.Getenv("OCI_RUNTIME")
		if bins["oci_runtime_path"] == "" {
			bins["oci_runtime_path"] = defaultRuntime[runtime.GOOS]
		}

		if bins["oci_runtime_path"] != "" {
			bins["garden_bin_path"], err = gexec.Build("code.cloudfoundry.org/guardian/cmd/guardian", "-tags", "daemon", "-race", "-ldflags", "-extldflags '-static'")
			Expect(err).NotTo(HaveOccurred())

			bins["dadoo_bin_bin_bin"], err = gexec.Build("code.cloudfoundry.org/guardian/cmd/dadoo")
			Expect(err).NotTo(HaveOccurred())

			bins["init_bin_path"], err = gexec.Build("code.cloudfoundry.org/guardian/cmd/init")
			Expect(err).NotTo(HaveOccurred())

			bins["inspector-garden_bin_path"], err = gexec.Build("code.cloudfoundry.org/guardian/cmd/inspector-garden")
			Expect(err).NotTo(HaveOccurred())

			bins["test_net_plugin_bin_path"], err = gexec.Build("code.cloudfoundry.org/guardian/gqt/cmd/networkplugin")
			Expect(err).NotTo(HaveOccurred())

			bins["test_image_plugin_bin_path"], err = gexec.Build("code.cloudfoundry.org/guardian/gqt/cmd/fake_image_plugin")
			Expect(err).NotTo(HaveOccurred())

			bins["containerd_shim_bin_path"], err = gexec.Build("github.com/docker/containerd/cmd/containerd-shim")
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command("make")
			cmd.Dir = "../rundmc/nstar"
			cmd.Stdout = GinkgoWriter
			cmd.Stderr = GinkgoWriter
			Expect(cmd.Run()).To(Succeed())
			bins["nstar_bin_path"] = "../rundmc/nstar/nstar"
		}

		data, err := json.Marshal(bins)
		Expect(err).NotTo(HaveOccurred())

		return data
	}, func(data []byte) {
		bins := make(map[string]string)
		Expect(json.Unmarshal(data, &bins)).To(Succeed())

		ociRuntimeBin = bins["oci_runtime_path"]
		gardenBin = bins["garden_bin_path"]
		nstarBin = bins["nstar_bin_path"]
		dadooBin = bins["dadoo_bin_bin_bin"]
		testImagePluginBin = bins["test_image_plugin_bin_path"]
		initBin = bins["init_bin_path"]
		inspectorGardenBin = bins["inspector-garden_bin_path"]
		testNetPluginBin = bins["test_net_plugin_bin_path"]
		containerdShimBin = bins["containerd_shim_bin_path"]

		tarBin = os.Getenv("GARDEN_TAR_PATH")
	})

	BeforeEach(func() {
		if ociRuntimeBin == "" {
			Skip("No OCI Runtime for Platform: " + runtime.GOOS)
		}

		if os.Getenv("GARDEN_TEST_ROOTFS") == "" {
			Skip("No Garden RootFS")
		}

		// chmod all the artifacts
		Expect(os.Chmod(filepath.Join(initBin, "..", ".."), 0755)).To(Succeed())
		filepath.Walk(filepath.Join(initBin, "..", ".."), func(path string, info os.FileInfo, err error) error {
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(path, 0755)).To(Succeed())
			return nil
		})
	})

	SetDefaultEventuallyTimeout(5 * time.Second)
	RunSpecs(t, "GQT Suite")
}

func startGarden(argv ...string) *runner.RunningGarden {
	rootfs := os.Getenv("GARDEN_TEST_ROOTFS")
	return runner.Start(gardenBin, initBin, nstarBin, containerdShimBin, testImagePluginBin, rootfs, tarBin, argv...)
}

func restartGarden(client *runner.RunningGarden, argv ...string) {
	Expect(client.Ping()).To(Succeed(), "tried to restart garden while it was not running")
	Expect(client.Stop()).To(Succeed())
	client = startGarden(argv...)
}

func startGardenWithoutDefaultRootfs(argv ...string) *runner.RunningGarden {
	return runner.Start(gardenBin, initBin, nstarBin, dadooBin, testImagePluginBin, "", tarBin, argv...)
}
