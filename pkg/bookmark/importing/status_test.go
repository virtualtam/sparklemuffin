// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "testing"

func TestAddStatusSummary(t *testing.T) {
	cases := []struct {
		tname  string
		status Status
		want   string
	}{
		{
			tname: "default status",
			want:  "0 new, 0 skipped, 0 invalid",
		},
		{
			tname: "import, keep existing",
			status: Status{
				NewOrUpdated: 17,
				Skipped:      3,
				Invalid:      4,
			},
			want: "17 new, 3 skipped, 4 invalid",
		},
		{
			tname: "import, overwrite existing",
			status: Status{
				overwriteExisting: true,
				NewOrUpdated:      17,
				Skipped:           0,
				Invalid:           4,
			},
			want: "17 new or updated, 0 skipped, 4 invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := tc.status.Summary()

			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
