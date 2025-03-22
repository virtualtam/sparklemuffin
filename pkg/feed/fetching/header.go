// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package fetching

import "time"

const (
	// The "ETag" field in a response provides the current entity tag for the
	// selected representation, as determined at the conclusion of handling the request.
	HeaderEntityTag string = "ETag"

	// The "If-None-Match" header field makes the request method conditional
	// on a recipient cache or origin server either not having any current representation
	// of the target resource, when the field value is "*", or having a selected
	// representation with an entity tag that does not match any of those listed
	// in the field value.
	HeaderIfNoneMatch string = "If-None-Match"

	// The "Last-Modified" header field in a response provides a timestamp indicating
	// the date and time at which the origin server believes the selected representation
	// was last modified, as determined at the conclusion of handling the request.
	HeaderLastModified string = "Last-Modified"

	// The "If-Modified-Since" header field makes a GET or HEAD request method
	// conditional on the selected representation's modification date being more recent
	// than the date provided in the field value. Transfer of the selected representation's
	// data is avoided if that data has not changed.
	HeaderIfModifiedSince string = "If-Modified-Since"
)

var (
	locationGMT = mustLoadTimeLocationGMT()
)

func mustLoadTimeLocationGMT() *time.Location {
	location, err := time.LoadLocation("GMT")
	if err != nil {
		panic(err)
	}

	return location
}

func formatLastModified(lastModified time.Time) string {
	return lastModified.In(locationGMT).Format(time.RFC1123)
}

func parseLastModified(lastModifiedStr string) time.Time {
	lastModified, err := time.ParseInLocation(time.RFC1123, lastModifiedStr, locationGMT)
	if err != nil {
		return time.Time{}
	}

	return lastModified
}
