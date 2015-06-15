package signalfx

/*
TODO(jrubin)
Package sfxmetric provides the types used for interacting with sfxreporter.

The fundamental type is Metric which is suggested for client use.

Additionally, Bucket is provided to track multiple aspects of a single metric
and Metrics is a container for storing multiple Metric objects.

It also provides a simple interface that can be used so that Metrics may
provide values via a callback at the time of the Report. In this way, pointer
values, or function values that may change are automatically kept up-to-date
without needing to manually set them.
*/
