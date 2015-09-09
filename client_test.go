package signalfx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
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

func TestClient(t *testing.T) {
	pdps := sfxproto.NewDataPoints(1).Add(&sfxproto.DataPoint{
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

	Convey("Testing Client", t, func(c C) {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			c.So(r.Header.Get(TokenHeader), ShouldEqual, "abcdefg")

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

		ts := httptest.NewServer(http.HandlerFunc(testHandler))
		defer ts.Close()

		config := NewConfig()
		config.AuthToken = "abcdefg"
		config.URL = ts.URL

		client := NewClient(config)
		So(client, ShouldNotBeNil)

		Convey("submit should work", func() {
			err := client.Submit(nil, pdps)
			So(err, ShouldBeNil)

			err = client.Submit(context.Background(), pdps)
			So(err, ShouldBeNil)
		})

		Convey("submit should handle a previously canceled context", func() {
			ctx, cancelF := context.WithCancel(context.Background())
			cancelF()
			<-ctx.Done()

			err := client.Submit(ctx, pdps)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "context canceled")
		})

		Convey("submit should handle an 'in-flight' context cancellation", func() {
			ctx, cancelF := context.WithCancel(context.Background())
			go cancelF()
			err := client.Submit(ctx, pdps)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "context canceled")
		})

		Convey("submit should handle various server errors properly", func() {
			mode = errStatusMode
			err := client.Submit(context.Background(), pdps)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `"OK": invalid status code: 404`)
			So(err, ShouldResemble, &ErrStatus{[]byte(`"OK"`), http.StatusNotFound})

			mode = errJSONMode
			err = client.Submit(context.Background(), pdps)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "OK")
			So(err, ShouldResemble, &ErrJSON{[]byte("OK")})

			mode = errInvalidBodyMode
			err = client.Submit(context.Background(), pdps)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "NOT OK")
			So(err, ShouldResemble, &ErrInvalidBody{"NOT OK"})

			pdps = nil
			err = client.Submit(context.Background(), pdps)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrMarshal(errors.New("no data to marshal")))

			ctx := context.Background()
			ctx, cancelFunc := context.WithCancel(ctx)
			// force a cancel on a non-cancellable request
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cancelFunc()
				w.Write([]byte(`"OK"`))
			}))
			defer ts.Close()
			config := config.Clone()
			config.URL = ts.URL
			// doesn't really matter that this is a
			// transportWrapper, just that
			// transportWrapper isn't cancellable
			client = NewClient(config)
			tw := transportWrapper{wrapped: client.tr}
			client.client = &http.Client{Transport: &tw}
			client.tr = &tw // can the SignalFX client just interrogate the HTTP client?
			pdps = sfxproto.NewDataPoints(1)
			metric := "foobar"
			pdps.Add(&sfxproto.DataPoint{Metric: &metric})
			err = client.Submit(ctx, pdps)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrContext(errors.New("context canceled")))
		})
	})
}
