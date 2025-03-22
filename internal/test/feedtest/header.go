// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feedtest

import "time"

const (
	headerEntityTag   string = "ETag"
	headerIfNoneMatch string = "If-None-Match"

	headerLastModified    string = "Last-Modified"
	headerIfModifiedSince string = "If-Modified-Since"
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
