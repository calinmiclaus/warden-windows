//package warden_windows_test
package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWardenWindows(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WardenWindows Suite")
}
