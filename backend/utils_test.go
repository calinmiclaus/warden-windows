package backend_test

import (
	. "github.com/UhuruSoftware/warden-windows/backend"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
)

var _ = Describe("Utils", func() {
	var rootPath string

	BeforeEach(func() {
		rootPath = "c:\\root"
	})

	Describe("adaptPathForPrison", func() {
		Context("when input path is absolute", func() {
			initialPath := "c:\\file1"

			It("return the input path", func() {
				adaptedPath := AdaptPathForPrison(rootPath, initialPath)
				Expect(adaptedPath).To(Equal(initialPath))
			})
		})
		Context("when input path is relative", func() {
			It("return a combined path from a path with backslash", func() {
				initialPath := "\\folder1\\file1"
				adaptedPath := AdaptPathForPrison(rootPath, initialPath)

				Expect(filepath.Clean(adaptedPath)).To(Equal(filepath.Clean("c:\\root\\folder1\\file1")))
			})

			It("return a combined path", func() {
				initialPath := "folder1\\file1"
				adaptedPath := AdaptPathForPrison(rootPath, initialPath)

				Expect(filepath.Clean(adaptedPath)).To(Equal(filepath.Clean("c:\\root\\folder1\\file1")))
			})
		})

		Context("when input path is unix-like", func() {
			It("return a combined path from a path with starting slash", func() {
				initialPath := "/tmp/app"
				adaptedPath := AdaptPathForPrison(rootPath, initialPath)

				Expect(filepath.Clean(adaptedPath)).To(Equal(filepath.Clean("c:\\root\\tmp\\app")))
			})

			It("return a combined path", func() {
				initialPath := "tmp/app"
				adaptedPath := AdaptPathForPrison(rootPath, initialPath)

				Expect(filepath.Clean(adaptedPath)).To(Equal(filepath.Clean("c:\\root\\tmp\\app")))
			})
		})
	})
})
