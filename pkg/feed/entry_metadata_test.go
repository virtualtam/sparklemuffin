// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import "testing"

func assertEntriesMetadataEqual(t *testing.T, gotEntriesMetadata, wantEntriesMetadata []EntryMetadata) {
	t.Helper()

	if len(gotEntriesMetadata) != len(wantEntriesMetadata) {
		t.Fatalf("want %d entries, got %d", len(wantEntriesMetadata), len(gotEntriesMetadata))
	}

	for i, wantEntryMetadata := range wantEntriesMetadata {
		gotEntryMetadata := gotEntriesMetadata[i]

		if gotEntryMetadata.UserUUID != wantEntryMetadata.UserUUID {
			t.Errorf("want UserUUID %q, got %q", wantEntryMetadata.UserUUID, gotEntryMetadata.UserUUID)
		}
		if gotEntryMetadata.EntryUID != wantEntryMetadata.EntryUID {
			t.Errorf("want EntryUID %q, got %q", wantEntryMetadata.EntryUID, gotEntryMetadata.EntryUID)
		}
		if gotEntryMetadata.Read != wantEntryMetadata.Read {
			t.Errorf("want Read %t, got %t", wantEntryMetadata.Read, gotEntryMetadata.Read)
		}
	}
}
