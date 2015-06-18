package signalfx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gogo/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

type handlerMode int

const (
	defaultMode handlerMode = iota
	errStatusMode
	errJSONMode
	errInvalidBodyMode
)

var (
	mode = defaultMode
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	switch mode {
	default:
		w.Write([]byte(`"OK"`))
	case errStatusMode:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"OK"`))
	case errJSONMode:
		w.Write([]byte("OK"))
	case errInvalidBodyMode:
		w.Write([]byte(`"NOT OK"`))
	}
}

func TestClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(testHandler))
	defer ts.Close()

	config := NewConfig()
	config.URL = ts.URL

	pdps := sfxproto.NewProtoDataPoints(1).Add(&sfxproto.ProtoDataPoint{
		Metric:     proto.String("TestClient"),
		MetricType: sfxproto.MetricType_COUNTER.Enum(),
		Dimensions: []*sfxproto.Dimension{{
			Key:   proto.String("test_client_0"),
			Value: proto.String("test_value_0"),
		}, {
			Key:   proto.String("test_client_1"),
			Value: proto.String("test_value_1"),
		}},
		Value: &sfxproto.Datum{
			IntValue: proto.Int64(7),
		},
	})

	Convey("Testing Client", t, func() {
		c := NewClient(config)
		So(c, ShouldNotBeNil)

		err := c.Submit(nil, pdps)
		So(err, ShouldBeNil)

		ctx, cancelF := context.WithCancel(context.Background())
		err = c.Submit(ctx, pdps)
		So(err, ShouldBeNil)

		cancelF()
		err = c.Submit(ctx, pdps)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "context canceled")

		ctx, cancelF = context.WithCancel(context.Background())
		go cancelF() // get an "in-flight" cancellation
		err = c.Submit(ctx, pdps)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "context canceled")

		mode = errStatusMode
		err = c.Submit(context.Background(), pdps)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, `"OK": invalid status code: 404`)
		So(err, ShouldResemble, &ErrStatus{[]byte(`"OK"`), http.StatusNotFound})

		mode = errJSONMode
		err = c.Submit(context.Background(), pdps)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "OK")
		So(err, ShouldResemble, &ErrJSON{[]byte("OK")})

		mode = errInvalidBodyMode
		err = c.Submit(context.Background(), pdps)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "NOT OK")
		So(err, ShouldResemble, &ErrInvalidBody{"NOT OK"})
	})
}
