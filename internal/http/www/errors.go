// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package www

import "errors"

var (
	ErrServerCSRFKeyRequired         = errors.New("server: csrf key required")
	ErrServerMetricsPrefixRequired   = errors.New("server: metrics prefix required")
	ErrServerMetricsRegistryRequired = errors.New("server: metrics registry required")
	ErrServerPublicURLRequired       = errors.New("server: public url required")

	ErrServerBookmarkServiceRequired          = errors.New("server: bookmark service required")
	ErrServerBookmarkExportingServiceRequired = errors.New("server: bookmark exporting service required")
	ErrServerBookmarkImportingServiceRequired = errors.New("server: bookmark importing service required")
	ErrServerBookmarkQueryingServiceRequired  = errors.New("server: bookmark querying service required")

	ErrServerFeedServiceRequired          = errors.New("server: feed service required")
	ErrServerFeedExportingServiceRequired = errors.New("server: feed exporting service required")
	ErrServerFeedImportingServiceRequired = errors.New("server: feed importing service required")
	ErrServerFeedQueryingServiceRequired  = errors.New("server: feed querying service required")

	ErrServerSessionServiceRequired = errors.New("server: session service required")
	ErrServerUserServiceRequired    = errors.New("server: user service required")
)
