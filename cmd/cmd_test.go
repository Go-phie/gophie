package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCmd(t *testing.T) {
	Convey("CLI [Test listing engines]\n", t, func() {
		args := []string{"engines", "list"}
		rootCmd.SetArgs(args)
		So(rootCmd.Execute(), ShouldBeNil)
	})

	Convey("CLI [Test current version]\n", t, func() {
		args := []string{"version"}
		rootCmd.SetArgs(args)
		So(rootCmd.Execute(), ShouldBeNil)
	})

	Convey("CLI [Test clearing cache]\n", t, func() {
		args := []string{"clear-cache"}
		rootCmd.SetArgs(args)
		So(rootCmd.Execute(), ShouldBeNil)
	})

}
