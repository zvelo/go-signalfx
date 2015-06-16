package signalfx

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestError(t *testing.T) {
	Convey("Testing errors", t, func() {
		var err error

		err = ErrStatus{[]byte("body"), 404}
		So(err.Error(), ShouldEqual, "body: invalid status code: 404")

		err = ErrJSON{[]byte("body")}
		So(err.Error(), ShouldEqual, "body")

		err = ErrInvalidBody{"body"}
		So(err.Error(), ShouldEqual, "body")
	})
}
