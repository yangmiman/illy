package fs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"provisioner/fs"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FS", func() {
	var (
		f       *fs.FS
		tempDir string
	)

	BeforeEach(func() {
		f = &fs.FS{}
		var err error
		tempDir, err = ioutil.TempDir("", "pcfdev-fs")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("#Mkdir", func() {
		Context("when the directory does not exist", func() {
			It("should create the directory", func() {
				Expect(f.Mkdir(filepath.Join(tempDir, "some-dir"))).To(Succeed())
				_, err := os.Stat(filepath.Join(tempDir, "some-dir"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the directory already exists", func() {
			BeforeEach(func() {
				Expect(os.Mkdir(filepath.Join(tempDir, "some-dir"), 0755)).To(Succeed())
			})

			It("should do nothing", func() {
				Expect(f.Mkdir(filepath.Join(tempDir, "some-dir"))).To(Succeed())
				_, err := os.Stat(filepath.Join(tempDir, "some-dir"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("#Write", func() {
		Context("when path is valid", func() {
			It("should create a file with path and writes contents", func() {
				readCloser := ioutil.NopCloser(strings.NewReader("some-contents"))
				Expect(f.Write(filepath.Join(tempDir, "some-file"), readCloser, os.FileMode(fs.FileModeRootReadWrite))).To(Succeed())
				data, err := ioutil.ReadFile(filepath.Join(tempDir, "some-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("some-contents"))
			})
		})

		Context("when file exists already", func() {
			BeforeEach(func() {
				Expect(f.Write(filepath.Join(tempDir, "some-file"), ioutil.NopCloser(strings.NewReader("some-content-that-is-really-long")), os.FileMode(fs.FileModeRootReadWrite))).To(Succeed())
			})

			It("should overwrite the file", func() {
				readCloser := ioutil.NopCloser(strings.NewReader("some-other-contents"))
				Expect(f.Write(filepath.Join(tempDir, "some-file"), readCloser, os.FileMode(fs.FileModeRootReadWrite))).To(Succeed())
				data, err := ioutil.ReadFile(filepath.Join(tempDir, "some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(data)).To(Equal("some-other-contents"))
			})
		})

		Context("when path is invalid", func() {
			It("should return an error", func() {
				Expect(f.Write(filepath.Join("some-bad-dir", "some-other-file"), nil, os.FileMode(fs.FileModeRootReadWrite))).To(MatchError(ContainSubstring("failed to open file:")))
			})
		})
	})

	Describe("#Read", func() {
		Context("when the file exists", func() {
			It("should return the contents the file", func() {
				Expect(ioutil.WriteFile(filepath.Join(tempDir, "some-file"), []byte("some-contents"), 0644)).To(Succeed())
				Expect(f.Read(filepath.Join(tempDir, "some-file"))).To(Equal([]byte("some-contents")))
			})
		})
	})

	Describe("#Exists", func() {
		Context("when the file exists", func() {
			BeforeEach(func() {
				_, err := os.Create(filepath.Join(tempDir, "some-file"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return true", func() {
				Expect(f.Exists(filepath.Join(tempDir, "some-file"))).To(BeTrue())
			})
		})

		Context("when the file does not exist", func() {
			It("should return false", func() {
				Expect(f.Exists(filepath.Join(tempDir, "some-bad-file"))).To(BeFalse())
			})
		})
	})
})
