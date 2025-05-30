package bookmark

import "testing"

func AssertBookmarkEquals(t *testing.T, got, want Bookmark) {
	t.Helper()

	if got.URL != want.URL {
		t.Errorf("want URL %q, got %q", want.URL, got.URL)
	}

	if got.Title != want.Title {
		t.Errorf("want Title %q, got %q", want.Title, got.Title)
	}

	if got.Description != want.Description {
		t.Errorf("want Description %q, got %q", want.Description, got.Description)
	}

	if got.Private != want.Private {
		t.Errorf("want Private %t, got %t", want.Private, got.Private)
	}

	if len(got.Tags) != len(want.Tags) {
		t.Fatalf("want %d tags, got %d", len(want.Tags), len(got.Tags))
	}

	for i, wantTag := range want.Tags {
		if got.Tags[i] != wantTag {
			t.Errorf("want tag %d Name %q, got %q", i, wantTag, got.Tags[i])
		}
	}
}
