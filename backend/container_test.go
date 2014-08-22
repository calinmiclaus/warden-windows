package backend_test

import (
	. "github.com/UhuruSoftware/warden-windows/backend"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Container", func() {
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

				//cri, err := PrisonClient.CreateContainerRunInfo()

				//Expect(err).ShouldNot(HaveOccurred())
				//Expect(cri).ShouldNot(BeNil())

				//err = cri.Release()
				//Expect(err).ShouldNot(HaveOccurred())
			})
		})
		Context("when input path is relative", func() {

		})
	})
})
