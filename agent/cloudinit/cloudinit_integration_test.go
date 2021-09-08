package cloudinit_test

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/cluster-api-provider-byoh/agent/cloudinit"
	"github.com/vmware-tanzu/cluster-api-provider-byoh/common"
)

var (
	workDir        string
	err            error
	scriptExecutor cloudinit.ScriptExecutor
)

var _ = Describe("CloudinitIntegration", func() {
	BeforeEach(func() {
		workDir, err = ioutil.TempDir("", "host-agent-ut")
		Expect(err).NotTo(HaveOccurred())

		scriptExecutor = cloudinit.ScriptExecutor{
			WriteFilesExecutor: cloudinit.FileWriter{},
			RunCmdExecutor:     cloudinit.CmdRunner{}}
	})

	It("should be able to write files and execute commands", func() {
		fileName := path.Join(workDir, "file-1.txt")
		fileOriginContent := "some-content-1"
		fileNewContent := " run cmd"

		cloudInitScript := fmt.Sprintf(`write_files:
- path: %s
content: %s
runCmd:
- echo -n '%s' > %s`, fileName, fileOriginContent, fileNewContent, fileName)

		err = scriptExecutor.Execute(cloudInitScript)
		Expect(err).ToNot(HaveOccurred())

		fileContents, errFileContents := ioutil.ReadFile(fileName)
		Expect(errFileContents).ToNot(HaveOccurred())
		Expect(string(fileContents)).To(Equal(fileNewContent))
	})

	It("should be able to write files with the correct permissions and in append mode", func() {
		fileName := path.Join(workDir, "file-2.txt")
		fileOriginContent := "some-content-2"
		fileAppendContent := "some-content-append-2"
		filePermission := 0777
		isAppend := true

		err = ioutil.WriteFile(fileName, []byte(fileOriginContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		cloudInitScript := fmt.Sprintf(`write_files:
- path: %s
  permissions: '%s'
  content: %s
  append: %v`, fileName, strconv.FormatInt(int64(filePermission), 8), fileAppendContent, isAppend)

		err = scriptExecutor.Execute(cloudInitScript)
		Expect(err).ToNot(HaveOccurred())

		fileContents, errFileContents := ioutil.ReadFile(fileName)
		Expect(errFileContents).ToNot(HaveOccurred())
		Expect(string(fileContents)).To(Equal(fileOriginContent + fileAppendContent))

		stats, statErr := os.Stat(fileName)
		Expect(statErr).ToNot(HaveOccurred())
		Expect(stats.Mode()).To(Equal(fs.FileMode(filePermission)))
	})

	It("should be able to write encoded content", func() {
		fileName := path.Join(workDir, "file-3.txt")
		fileContent := "some-content-3"
		fileBase64Content := base64.StdEncoding.EncodeToString([]byte(fileContent))

		cloudInitScript := fmt.Sprintf(`write_files:
- path: %s
  content: %s
  encoding: base64`, fileName, fileBase64Content)

		err = scriptExecutor.Execute(cloudInitScript)
		Expect(err).ToNot(HaveOccurred())

		fileContents, err := ioutil.ReadFile(fileName)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(fileContents)).To(Equal(fileContent))
	})

	It("should be able to write gziped content", func() {
		fileName := path.Join(workDir, "file-4.txt")
		fileContent := "some-content-4"
		fileGzipContent, err := common.GzipData([]byte(fileContent))
		Expect(err).NotTo(HaveOccurred())
		fileGzipBase64Content := base64.StdEncoding.EncodeToString(fileGzipContent)

		cloudInitScript := fmt.Sprintf(`write_files:
- path: %s
  encoding: gzip+base64
  content: %s`, fileName, fileGzipBase64Content)

		err = scriptExecutor.Execute(cloudInitScript)
		Expect(err).ToNot(HaveOccurred())

		fileContents, err := ioutil.ReadFile(fileName)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(fileContents)).To(Equal(fileContent))
	})

	AfterEach(func() {
		err := os.RemoveAll(workDir)
		Expect(err).ToNot(HaveOccurred())
	})
})