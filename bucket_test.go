package signalfx

import (
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestBucket(t *testing.T) {
	Convey("Testing Bucket", t, func() {
		b := NewBucket("test", map[string]string{"c": "3"})
		So(b, ShouldNotBeNil)
		So(len(b.DataPoints()), ShouldEqual, 3)
		So(b.min, ShouldEqual, math.MaxInt64)
		So(b.max, ShouldEqual, math.MinInt64)

		Convey("metric naming should be correct", func() {
			So(b.Metric(), ShouldEqual, "test")
			b.SetMetric("new metric name")
			So(b.Metric(), ShouldEqual, "new metric name")
		})

		Convey("dimension operations should work", func() {
			So(b.Dimensions(), ShouldResemble, map[string]string{"c": "3"})

			b.SetDimension("a", "")
			So(b.Dimensions(), ShouldResemble, map[string]string{
				"c": "3",
			})

			b.SetDimension("", "1")
			So(b.Dimensions(), ShouldResemble, map[string]string{
				"c": "3",
			})

			b.SetDimension("", "")
			So(b.Dimensions(), ShouldResemble, map[string]string{
				"c": "3",
			})

			b.SetDimension("a", "1")
			So(b.Dimensions(), ShouldResemble, map[string]string{
				"a": "1",
				"c": "3",
			})

			dims := map[string]string{"b": "2"}
			b.SetDimensions(dims)
			So(b.Dimensions(), ShouldResemble, map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			})

			b.RemoveDimension("c")
			So(b.Dimensions(), ShouldResemble, map[string]string{
				"a": "1",
				"b": "2",
			})

			b.RemoveDimension("a", "b")
			So(b.Dimensions(), ShouldResemble, map[string]string{})
		})

		Convey("Clone/Equal should work correctly", func() {
			c := b.Clone()
			So(b, ShouldResemble, c)
			So(c.Equal(b), ShouldBeTrue)
			So(b.Equal(c), ShouldBeTrue)

			tmp := &Bucket{}
			So(b.Equal(tmp), ShouldBeFalse)
			So(tmp.Equal(b), ShouldBeFalse)

			So(c.Equal(tmp), ShouldBeFalse)
			So(tmp.Equal(c), ShouldBeFalse)

			c.sumOfSquares = b.sumOfSquares + 1
			So(c.Equal(b), ShouldBeFalse)
			So(b.Equal(c), ShouldBeFalse)

			c.sum = b.sum + 1
			So(c.Equal(b), ShouldBeFalse)
			So(b.Equal(c), ShouldBeFalse)

			c.max = b.max + 1
			So(c.Equal(b), ShouldBeFalse)
			So(b.Equal(c), ShouldBeFalse)

			c.min = b.min + 1
			So(c.Equal(b), ShouldBeFalse)
			So(b.Equal(c), ShouldBeFalse)

			c.count = b.count + 1
			So(c.Equal(b), ShouldBeFalse)
			So(b.Equal(c), ShouldBeFalse)

			c.dimensions = map[string]string{}
			So(c.Equal(b), ShouldBeFalse)
			So(b.Equal(c), ShouldBeFalse)
		})

		Convey("data handling should be correct", func() {
			b.Add(1)
			b.Add(2)
			b.Add(3)

			So(b.Max(), ShouldEqual, 3)
			So(b.Min(), ShouldEqual, 1)
			So(b.Count(), ShouldEqual, 3)
			So(b.Sum(), ShouldEqual, 6)
			So(b.SumOfSquares(), ShouldEqual, 14)

			b.Add(0)
			b.Add(10)
			b.Add(1)

			So(b.Max(), ShouldEqual, 10)
			So(b.Min(), ShouldEqual, 0)

			b.Add(1)
			So(len(b.DataPoints()), ShouldEqual, 5)
		})

		Convey("rollup dimensions should be added", func() {
			b.Add(1)
			b.Add(2)
			b.Add(3)

			datapoints := b.DataPoints()
			So(len(datapoints), ShouldEqual, 5)
			So(datapoints[0].Dimensions["rollup"], ShouldResemble, "min")
			So(datapoints[1].Dimensions["rollup"], ShouldResemble, "max")
			So(datapoints[2].Dimensions["rollup"], ShouldResemble, "count")
			So(datapoints[3].Dimensions["rollup"], ShouldResemble, "sum")
			So(datapoints[4].Dimensions["rollup"], ShouldResemble, "sumofsquares")
		})

		Convey("bucket datapoints should be sensible", func() {
			t := time.Now()
			dps := b.DataPoints()
			So(len(dps), ShouldBeGreaterThan, 0)
			So(dps[0].Timestamp.After(t), ShouldBeTrue)
		})

		Convey("disabling datapoints should work", func() {
			b.Disable(BucketMetricCount)

			b.Add(1)
			b.Add(2)
			b.Add(3)

			So(b.Count(), ShouldEqual, 3)
			So(len(b.DataPoints()), ShouldEqual, 4)
		})
	})
}

func ExampleBucket() {
	reporter := NewReporter(NewConfig(), map[string]string{
		"test_dimension0": "value0",
		"test_dimension1": "value1",
	})

	bucket := reporter.NewBucket("TestBucket", map[string]string{
		"test_bucket_dimension0": "bucket0",
		"test_bucket_dimension1": "bucket1",
	})

	bucket.Add(5)
	bucket.Add(9)

	fmt.Printf("Metric: %s\n", bucket.Metric())
	fmt.Printf("Count: %d\n", bucket.Count())
	fmt.Printf("Min: %d\n", bucket.Min())
	fmt.Printf("Max: %d\n", bucket.Max())
	fmt.Printf("Sum: %d\n", bucket.Sum())
	fmt.Printf("SumOfSquares: %d\n", bucket.SumOfSquares())

	dps, err := reporter.Report(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Metrics:", len(dps))
	}

	// Output:
	// Metric: TestBucket
	// Count: 2
	// Min: 5
	// Max: 9
	// Sum: 14
	// SumOfSquares: 106
	// Metrics: 5
}
