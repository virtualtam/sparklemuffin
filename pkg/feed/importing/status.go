// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "fmt"

type Status struct {
	Categories    StatusCount
	Feeds         StatusCount
	Subscriptions StatusCount
}

func (s *Status) AdminSummary() string {
	return fmt.Sprintf(
		"%d categories (%d new), %d feeds (%d new), %d subscriptions (%d new)",
		s.Categories.Total,
		s.Categories.Created,
		s.Feeds.Total,
		s.Feeds.Created,
		s.Subscriptions.Total,
		s.Subscriptions.Created,
	)
}

func (s *Status) UserSummary() string {
	return fmt.Sprintf(
		"%d categories (%d new), %d subscriptions (%d new)",
		s.Categories.Total,
		s.Categories.Created,
		s.Subscriptions.Total,
		s.Subscriptions.Created,
	)
}

type StatusCount struct {
	Total   uint
	Created uint
}

func (sc *StatusCount) Inc(created bool) {
	sc.Total++

	if created {
		sc.Created++
	}
}
