package bookmark

import (
	"errors"
	"testing"
)

func TestServiceAdd(t *testing.T) {
	cases := []struct {
		tname    string
		bookmark Bookmark
		wantErr  error
	}{
		{
			tname:   "empty bookmark",
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname: "missing user UUID",
			bookmark: Bookmark{
				URL:   "https://domain.tld",
				Title: "Example Domain",
			},
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname: "empty URL",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "empty (whitespace) URL",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "   ",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "unparseable URL",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      ":/dmn",
			},
			wantErr: ErrURLInvalid,
		},
		{
			tname: "empty title",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "empty (whitespace) title",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "    ",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "add bookmark",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
		},
		{
			tname: "add bookmark with description",
			bookmark: Bookmark{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:         "https://domain.tld",
				Title:       "Example Domain",
				Description: "Hello,\nThis bookmark has a longer description!",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{}
			s := NewService(r)

			err := s.Add(tc.bookmark)

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
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []Bookmark
		bookmark            Bookmark
		wantErr             error
	}{
		{
			tname:   "empty bookmark",
			wantErr: ErrUIDRequired,
		},
		{
			tname: "invalid UID",
			bookmark: Bookmark{
				UID: "12345",
			},
			wantErr: ErrUIDInvalid,
		},
		{
			tname: "missing user UUID",
			bookmark: Bookmark{
				UID:   "27L4DoEZaRASKhQKygRCrvVAwkr",
				URL:   "https://domain.tld",
				Title: "Example Domain",
			},
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname: "empty URL",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "empty (whitespace) URL",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "   ",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "unparseable URL",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      ":/dmn",
			},
			wantErr: ErrURLInvalid,
		},
		{
			tname: "empty title",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "empty (whitespace) title",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "    ",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "not found",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
			wantErr: ErrNotFound,
		},
		{
			tname: "update bookmark",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Title:    "Example Doma",
				},
			},
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
		},
		{
			tname: "update bookmark with description",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Title:    "Example Doma",
				},
			},
			bookmark: Bookmark{
				UID:         "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:         "https://domain.tld",
				Title:       "Example Domain",
				Description: "Hello,\nThis bookmark has a longer description!",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			err := s.Update(tc.bookmark)

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
		})
	}
}

func TestServiceDelete(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []Bookmark
		userUUID            string
		uid                 string
		wantErr             error
	}{
		{
			tname:   "empty UID",
			wantErr: ErrUIDRequired,
		},
		{
			tname:   "invalid UID",
			uid:     "invalid",
			wantErr: ErrUIDInvalid,
		},
		{
			tname:   "empty user UUID",
			uid:     "27L5erU5VNJzIGY1uPUqzLkc9zV",
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname:    "not found",
			uid:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
			userUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
			wantErr:  ErrNotFound,
		},
		{
			tname: "delete bookmark",
			repositoryBookmarks: []Bookmark{
				{
					UserUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
					UID:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
					URL:      "https://domain.tld",
					Title:    "Test Domain",
				},
			},
			uid:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
			userUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			err := s.Delete(tc.userUUID, tc.uid)

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
		})
	}
}
