package properties_test

import (
	"github.com/cloudfoundry-incubator/guardian/properties"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Properties", func() {
	var (
		propertyManager *properties.Manager
	)

	Describe("Manager", func() {
		BeforeEach(func() {
			propertyManager = properties.NewManager()

			err := propertyManager.CreateKeySpace("handle")
			Expect(err).NotTo(HaveOccurred())

			err = propertyManager.Set("handle", "name", "value")
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("CreateKeySpace", func() {
			It("does not reinitialize key space if present", func() {
				val, err := propertyManager.Get("handle", "name")
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal("value"))

				err = propertyManager.CreateKeySpace("handle")
				Expect(err).NotTo(HaveOccurred())

				val, err = propertyManager.Get("handle", "name")
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal("value"))
			})
		})

		Describe("DestroyKeySpace", func() {
			It("removes key space", func() {
				err := propertyManager.DestroyKeySpace("handle")
				Expect(err).NotTo(HaveOccurred())

				err = propertyManager.DestroyKeySpace("handle")
				Expect(err).To(MatchError(properties.NoSuchKeySpaceError{"No such key space: handle"}))
			})
		})

		Describe("All", func() {
			It("returns the properties", func() {
				props, err := propertyManager.All("handle")
				Expect(err).NotTo(HaveOccurred())

				Expect(props).To(HaveLen(1))
				Expect(props).To(HaveKeyWithValue("name", "value"))
			})
		})

		Describe("Get", func() {
			It("returns a specific property when passed a name", func() {
				property, err := propertyManager.Get("handle", "name")

				Expect(err).NotTo(HaveOccurred())
				Expect(property).To(Equal("value"))
			})
		})

		Describe("Remove", func() {
			It("removes properties", func() {
				props, err := propertyManager.All("handle")
				Expect(err).NotTo(HaveOccurred())

				Expect(props).To(HaveLen(1))

				err = propertyManager.Remove("handle", "name")
				Expect(err).NotTo(HaveOccurred())

				_, err = propertyManager.Get("handle", "name")
				Expect(err).To(MatchError(properties.NoSuchPropertyError{"No such property: name"}))
			})
		})

		Describe("Set", func() {
			Context("when the property already exists", func() {
				It("updates the property value", func() {
					err := propertyManager.Set("handle", "name", "some-other-value")
					Expect(err).NotTo(HaveOccurred())

					props, err := propertyManager.All("handle")
					Expect(err).NotTo(HaveOccurred())
					Expect(props).To(HaveKeyWithValue("name", "some-other-value"))
				})
			})
		})

		Context("when attempting to remove a property that doesn't exist", func() {
			It("returns a NoSuchPropertyError", func() {
				err := propertyManager.Remove("handle", "missing")
				Expect(err).To(MatchError(properties.NoSuchPropertyError{"No such property: missing"}))
			})
		})
	})
})
