// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import (
	"errors"
	"fmt"
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestServiceDeleteTag(t *testing.T) {
	cases := []struct {
		tname          string
		repositoryTags []Tag
		tagDeleteQuery TagDeleteQuery
		wantErr        error
		wantTags       []Tag
	}{
		// nominal cases
		{
			tname: "delete a tag",
			repositoryTags: []Tag{
				{UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "a"},
				{UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "delete-me"},
			},
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "delete-me",
			},
			wantTags: []Tag{
				{UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "a"},
			},
		},

		// edge cases
		{
			tname: "no tag with this name",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "unknown",
			},
		},
		{
			tname: "delete a tag does not affect other users",
			repositoryTags: []Tag{
				{UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "shared-name"},
				{UserUUID: "other-user-uuid", Name: "shared-name"},
			},
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "shared-name",
			},
			wantTags: []Tag{
				{UserUUID: "other-user-uuid", Name: "shared-name"},
			},
		},

		// error cases
		{
			tname: "name is empty",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "name is empty (whitespace)",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "     ",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "name contains whitespace (multiple values)",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "tag1   tag2",
			},
			wantErr: ErrTagNameContainsWhitespace,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: tc.repositoryTags,
			}
			s := NewService(r)

			err := s.DeleteTag(t.Context(), tc.tagDeleteQuery)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(r.Tags) != len(tc.wantTags) {
				t.Fatalf("want %d remaining tags, got %d", len(tc.wantTags), len(r.Tags))
			}

			for index, tag := range r.Tags {
				want := tc.wantTags[index]
				if tag.UserUUID != want.UserUUID || tag.Name != want.Name {
					t.Errorf("want tag %+v, got %+v", want, tag)
				}
			}
		})
	}
}

func TestServiceRenameTag(t *testing.T) {
	cases := []struct {
		tname             string
		repositoryTags    []Tag
		tagUpdateQuery    TagUpdateQuery
		mergeTagErr       error
		wantErr           error
		wantTag           Tag
		wantTags          []Tag
		wantMergeTagCalls []MergeTagCall
	}{
		// nominal cases
		{
			tname: "rename a tag",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
			},
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo", NewName: "bar"},
			wantTag:        Tag{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "bar"},
			wantTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "bar"},
			},
		},
		{
			tname: "rename a tag into an already existing tag merges them",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
				{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "bar"},
			},
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo", NewName: "bar"},
			wantTag:        Tag{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "bar"},
			wantTags: []Tag{
				{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "bar"},
			},
			wantMergeTagCalls: []MergeTagCall{
				{UserUUID: "user-1", OldTagUUID: "tag-a-uuid", NewTagUUID: "tag-b-uuid"},
			},
		},

		// edge cases
		{
			tname: "new name is the same as the current name",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
			},
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo", NewName: "foo"},
			wantTag:        Tag{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
			wantTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
			},
		},
		{
			tname:          "no tag with the current name",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "unknown", NewName: "bar"},
		},

		// error cases
		{
			tname:          "current name is empty",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1"},
			wantErr:        ErrTagNameRequired,
		},
		{
			tname:          "current name is empty (whitespace)",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "     "},
			wantErr:        ErrTagNameRequired,
		},
		{
			tname:          "current name contains whitespace (multiple values)",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "tag1   tag2"},
			wantErr:        ErrTagNameContainsWhitespace,
		},
		{
			tname:          "new name is empty",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo"},
			wantErr:        ErrTagNameRequired,
		},
		{
			tname:          "new name is empty (whitespace)",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo", NewName: "     "},
			wantErr:        ErrTagNameRequired,
		},
		{
			tname:          "new name contains whitespace (multiple values)",
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo", NewName: "tag2 tag3   tag4"},
			wantErr:        ErrTagNameContainsWhitespace,
		},
		{
			tname: "merge fails, both tags are preserved",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
				{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "bar"},
			},
			tagUpdateQuery: TagUpdateQuery{UserUUID: "user-1", CurrentName: "foo", NewName: "bar"},
			mergeTagErr:    errors.New("merge failed"),
			wantErr:        errors.New("merge failed"),
			wantTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "foo"},
				{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "bar"},
			},
			wantMergeTagCalls: []MergeTagCall{
				{UserUUID: "user-1", OldTagUUID: "tag-a-uuid", NewTagUUID: "tag-b-uuid"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags:        tc.repositoryTags,
				MergeTagErr: tc.mergeTagErr,
			}
			s := NewService(r)

			got, err := s.RenameTag(t.Context(), tc.tagUpdateQuery)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) && err.Error() != tc.wantErr.Error() {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.UUID != tc.wantTag.UUID || got.UserUUID != tc.wantTag.UserUUID || got.Name != tc.wantTag.Name {
				t.Errorf("want returned tag %+v, got %+v", tc.wantTag, got)
			}

			if len(r.Tags) != len(tc.wantTags) {
				t.Fatalf("want %d remaining tags, got %d", len(tc.wantTags), len(r.Tags))
			}

			for index, tag := range r.Tags {
				want := tc.wantTags[index]
				if tag.UUID != want.UUID || tag.UserUUID != want.UserUUID || tag.Name != want.Name {
					t.Errorf("want tag %+v, got %+v", want, tag)
				}
			}

			if len(r.MergeTagCalls) != len(tc.wantMergeTagCalls) {
				t.Fatalf("want %d MergeTag calls, got %d", len(tc.wantMergeTagCalls), len(r.MergeTagCalls))
			}

			for index, call := range r.MergeTagCalls {
				want := tc.wantMergeTagCalls[index]
				if call != want {
					t.Errorf("want MergeTag call %+v, got %+v", want, call)
				}
			}
		})
	}
}

func TestServiceAddTag(t *testing.T) {
	cases := []struct {
		tname          string
		repositoryTags []Tag
		userUUID       string
		name           string
		wantErr        error
		wantName       string
	}{
		// nominal cases
		{
			tname:    "add a tag",
			userUUID: "user-1",
			name:     "new",
			wantName: "new",
		},

		// edge cases
		{
			tname:    "name is trimmed",
			userUUID: "user-1",
			name:     "  new  ",
			wantName: "new",
		},

		// error cases
		{
			tname:   "user UUID is empty",
			name:    "new",
			wantErr: user.ErrUUIDRequired,
		},
		{
			tname:    "name is empty",
			userUUID: "user-1",
			wantErr:  ErrTagNameRequired,
		},
		{
			tname:    "name contains whitespace (multiple values)",
			userUUID: "user-1",
			name:     "tag1   tag2",
			wantErr:  ErrTagNameContainsWhitespace,
		},
		{
			tname: "name is already registered",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "existing"},
			},
			userUUID: "user-1",
			name:     "existing",
			wantErr:  ErrTagAlreadyRegistered,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: tc.repositoryTags,
			}
			s := NewService(r)

			got, err := s.AddTag(t.Context(), tc.userUUID, tc.name)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.UUID == "" {
				t.Errorf("want a non-empty UUID for tag %+v", got)
			}
			if got.UserUUID != tc.userUUID {
				t.Errorf("want UserUUID %q, got %q", tc.userUUID, got.UserUUID)
			}
			if got.Name != tc.wantName {
				t.Errorf("want tag name %q, got %q", tc.wantName, got.Name)
			}

			if len(r.Tags) != len(tc.repositoryTags)+1 {
				t.Errorf("want %d tags in the repository, got %d", len(tc.repositoryTags)+1, len(r.Tags))
			}
		})
	}
}

func TestServiceGetOrCreateTags(t *testing.T) {
	cases := []struct {
		tname          string
		repositoryTags []Tag
		userUUID       string
		names          []string
		wantErr        error
		wantNames      []string
		wantNewTags    int
	}{
		// nominal cases
		{
			tname: "resolve an existing tag and create a missing one",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "existing"},
			},
			userUUID:    "user-1",
			names:       []string{"existing", "new"},
			wantNames:   []string{"existing", "new"},
			wantNewTags: 1,
		},
		{
			tname: "resolve only existing tags creates nothing",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "a"},
				{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "b"},
			},
			userUUID:  "user-1",
			names:     []string{"a", "b"},
			wantNames: []string{"a", "b"},
		},

		// edge cases
		{
			tname:     "no names resolves to no tags",
			userUUID:  "user-1",
			wantNames: []string{},
		},
		{
			tname: "the same tag existing for another user is not reused",
			repositoryTags: []Tag{
				{UUID: "tag-a-uuid", UserUUID: "other-user-uuid", Name: "shared-name"},
			},
			userUUID:    "user-1",
			names:       []string{"shared-name"},
			wantNames:   []string{"shared-name"},
			wantNewTags: 1,
		},

		// error cases
		{
			tname:   "user UUID is empty",
			names:   []string{"a"},
			wantErr: user.ErrUUIDRequired,
		},
		{
			tname:    "name contains whitespace (multiple values)",
			userUUID: "user-1",
			names:    []string{"tag1   tag2"},
			wantErr:  ErrTagNameContainsWhitespace,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: tc.repositoryTags,
			}
			s := NewService(r)

			got, err := s.GetOrCreateTags(t.Context(), tc.userUUID, tc.names)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(got) != len(tc.wantNames) {
				t.Fatalf("want %d tags, got %d", len(tc.wantNames), len(got))
			}

			for index, tag := range got {
				if tag.UUID == "" {
					t.Errorf("want a non-empty UUID for tag %+v", tag)
				}
				if tag.UserUUID != tc.userUUID {
					t.Errorf("want UserUUID %q, got %q", tc.userUUID, tag.UserUUID)
				}
				if tag.Name != tc.wantNames[index] {
					t.Errorf("want tag name %q, got %q", tc.wantNames[index], tag.Name)
				}
			}

			if len(r.Tags) != len(tc.repositoryTags)+tc.wantNewTags {
				t.Errorf("want %d tags in the repository, got %d", len(tc.repositoryTags)+tc.wantNewTags, len(r.Tags))
			}
		})
	}
}

func TestServiceListTags(t *testing.T) {
	tags := make([]Tag, 3)
	for i := range tags {
		tags[i] = Tag{UUID: fmt.Sprintf("tag-%d-uuid", i), UserUUID: "user-1", Name: fmt.Sprintf("tag-%d", i)}
	}

	cases := []struct {
		tname          string
		repositoryTags []Tag
		number         uint
		wantErr        error
		wantNames      []string
		wantPageNumber uint
		wantTotalPages uint
	}{
		// nominal cases
		{
			tname:          "first (and only) page",
			repositoryTags: tags,
			number:         1,
			wantNames:      []string{"tag-0", "tag-1", "tag-2"},
			wantPageNumber: 1,
			wantTotalPages: 1,
		},

		// edge cases
		{
			tname:          "no tags",
			number:         1,
			wantNames:      []string{},
			wantPageNumber: 1,
			wantTotalPages: 1,
		},

		// error cases
		{
			tname:          "page number is zero",
			repositoryTags: tags,
			number:         0,
			wantErr:        paginate.ErrPageNumberOutOfBounds,
		},
		{
			tname:          "page number is out of bounds",
			repositoryTags: tags,
			number:         2,
			wantErr:        paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: tc.repositoryTags,
			}
			s := NewService(r)

			got, err := s.ListTags(t.Context(), "user-1", tc.number)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.PageNumber != tc.wantPageNumber {
				t.Errorf("want page number %d, got %d", tc.wantPageNumber, got.PageNumber)
			}
			if got.TotalPages != tc.wantTotalPages {
				t.Errorf("want %d total pages, got %d", tc.wantTotalPages, got.TotalPages)
			}

			if len(got.Tags) != len(tc.wantNames) {
				t.Fatalf("want %d tags, got %d", len(tc.wantNames), len(got.Tags))
			}
			for index, tag := range got.Tags {
				if tag.Name != tc.wantNames[index] {
					t.Errorf("want tag name %q, got %q", tc.wantNames[index], tag.Name)
				}
			}
		})
	}
}

func TestServiceSearchTags(t *testing.T) {
	repositoryTags := []Tag{
		{UUID: "tag-a-uuid", UserUUID: "user-1", Name: "golang"},
		{UUID: "tag-b-uuid", UserUUID: "user-1", Name: "gopher"},
		{UUID: "tag-c-uuid", UserUUID: "user-1", Name: "python"},
	}

	cases := []struct {
		tname       string
		searchTerms string
		number      uint
		wantErr     error
		wantNames   []string
	}{
		// nominal cases
		{
			tname:       "search matches a subset of tags",
			searchTerms: "go",
			number:      1,
			wantNames:   []string{"golang", "gopher"},
		},

		// edge cases
		{
			tname:       "search matches no tags",
			searchTerms: "rust",
			number:      1,
			wantNames:   []string{},
		},

		// error cases
		{
			tname:       "page number is zero",
			searchTerms: "go",
			number:      0,
			wantErr:     paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: repositoryTags,
			}
			s := NewService(r)

			got, err := s.SearchTags(t.Context(), "user-1", tc.searchTerms, tc.number)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.SearchTerms != tc.searchTerms {
				t.Errorf("want search terms %q, got %q", tc.searchTerms, got.SearchTerms)
			}

			if len(got.Tags) != len(tc.wantNames) {
				t.Fatalf("want %d tags, got %d", len(tc.wantNames), len(got.Tags))
			}
			for index, tag := range got.Tags {
				if tag.Name != tc.wantNames[index] {
					t.Errorf("want tag name %q, got %q", tc.wantNames[index], tag.Name)
				}
			}
		})
	}
}

func TestServiceAllTags(t *testing.T) {
	cases := []struct {
		tname          string
		repositoryTags []Tag
		wantNames      []string
	}{
		// nominal cases
		{
			tname: "several tags",
			repositoryTags: []Tag{
				{UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "golang"},
				{UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "rss"},
				{UserUUID: "other-user-uuid", Name: "python"},
			},
			wantNames: []string{"golang", "rss"},
		},

		// edge cases
		{
			tname:     "no tags",
			wantNames: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: tc.repositoryTags,
			}
			s := NewService(r)

			got, err := s.AllTags(t.Context(), "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096")
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(got) != len(tc.wantNames) {
				t.Fatalf("want %d tags, got %d", len(tc.wantNames), len(got))
			}
			for index, tag := range got {
				if tag.Name != tc.wantNames[index] {
					t.Errorf("want tag name %q, got %q", tc.wantNames[index], tag.Name)
				}
			}
		})
	}
}

func TestServiceTagByUUID(t *testing.T) {
	cases := []struct {
		tname          string
		repositoryTags []Tag
		userUUID       string
		tagUUID        string
		wantErr        error
		wantName       string
	}{
		// nominal cases
		{
			tname: "tag exists",
			repositoryTags: []Tag{
				{UUID: "d290f1ee-6c54-4b01-90e6-d701748f0851", UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096", Name: "existing"},
			},
			userUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			tagUUID:  "d290f1ee-6c54-4b01-90e6-d701748f0851",
			wantName: "existing",
		},

		// edge cases
		{
			tname:    "no tag with this UUID",
			userUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			tagUUID:  "d290f1ee-6c54-4b01-90e6-d701748f0851",
			wantErr:  ErrNotFound,
		},
		{
			tname: "tag exists for another user",
			repositoryTags: []Tag{
				{UUID: "d290f1ee-6c54-4b01-90e6-d701748f0851", UserUUID: "other-user-uuid", Name: "existing"},
			},
			userUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			tagUUID:  "d290f1ee-6c54-4b01-90e6-d701748f0851",
			wantErr:  ErrNotFound,
		},

		// error cases
		{
			tname:    "UUID is empty",
			userUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			wantErr:  ErrTagUUIDInvalid,
		},
		{
			tname:    "UUID is malformed",
			userUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			tagUUID:  "not-a-uuid",
			wantErr:  ErrTagUUIDInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Tags: tc.repositoryTags,
			}
			s := NewService(r)

			got, err := s.TagByUUID(t.Context(), tc.userUUID, tc.tagUUID)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if got.Name != tc.wantName {
				t.Errorf("want tag name %q, got %q", tc.wantName, got.Name)
			}
		})
	}
}
