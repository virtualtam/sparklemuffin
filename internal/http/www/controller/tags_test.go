// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var testTagsCtxUser = user.User{UUID: "user-1", NickName: "alice", DisplayName: "Alice"}

// newTestTagsController wires a tagsController against fake taxonomy,
// bookmark querying and feed querying repositories.
func newTestTagsController(tags []taxonomy.Tag, bookmarks []bookmark.Bookmark, subscriptions []feed.Subscription) tagsController {
	taxonomyRepo := &taxonomy.FakeRepository{Tags: tags}

	bookmarkQueryingRepo := &bookmarkquerying.FakeRepository{
		Bookmarks: bookmarks,
		Users:     []user.User{testTagsCtxUser},
	}

	feedQueryingRepo := &feedquerying.FakeRepository{
		Subscriptions: subscriptions,
	}

	return tagsController{
		taxonomyService:         taxonomy.NewService(taxonomyRepo),
		bookmarkQueryingService: bookmarkquerying.NewService(bookmarkQueryingRepo),
		feedQueryingService:     feedquerying.NewService(feedQueryingRepo),

		tagListView:   view.New("taxonomy/tag_list.gohtml"),
		tagEditView:   view.New("taxonomy/tag_edit.gohtml"),
		tagDeleteView: view.New("taxonomy/tag_delete.gohtml"),
	}
}

// newTagUUIDRequest builds a request against /tags/{uuid}/... carrying the
// given method, tag UUID URL param, and ctxUser in context.
func newTagUUIDRequest(t *testing.T, method string, tagUUID string, form url.Values, hxRequest bool) *http.Request {
	t.Helper()

	target := "/tags/" + tagUUID
	body := strings.NewReader("")
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	r := httptest.NewRequestWithContext(t.Context(), method, target, body)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if hxRequest {
		r.Header.Set("HX-Request", "true")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", tagUUID)

	ctx := httpcontext.WithUser(r.Context(), testTagsCtxUser)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return r.WithContext(ctx)
}

func TestHandleTagListView(t *testing.T) {
	fake := faker.New()
	golangUUID := fake.UUID().V4()
	rssUUID := fake.UUID().V4()

	tags := []taxonomy.Tag{
		{UUID: golangUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"},
		{UUID: rssUUID, UserUUID: testTagsCtxUser.UUID, Name: "rss"},
	}
	bookmarks := []bookmark.Bookmark{
		{UID: "b1", UserUUID: testTagsCtxUser.UUID, Title: "Bookmark 1", URL: "https://example1.tld", Tags: []string{"golang"}},
		{UID: "b2", UserUUID: testTagsCtxUser.UUID, Title: "Bookmark 2", URL: "https://example2.tld", Tags: []string{"golang"}},
	}
	subscriptions := []feed.Subscription{
		{UUID: "s1", UserUUID: testTagsCtxUser.UUID, Tags: []string{"rss"}},
	}

	tc := newTestTagsController(tags, bookmarks, subscriptions)

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tags", nil)
	ctx := httpcontext.WithUser(r.Context(), testTagsCtxUser)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	tc.handleTagListView()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "golang") {
		t.Errorf("want tag %q rendered, got:\n%s", "golang", body)
	}
	if !strings.Contains(body, "rss") {
		t.Errorf("want tag %q rendered, got:\n%s", "rss", body)
	}
	if !strings.Contains(body, `id="tag-row-`+golangUUID+`"`) {
		t.Errorf("want tag row identified by UUID, got:\n%s", body)
	}

	// golang (2 bookmarks) has a higher total usage count than rss (1 feed
	// subscription), so it must be rendered first.
	if gi, ri := strings.Index(body, "tag-row-"+golangUUID), strings.Index(body, "tag-row-"+rssUUID); gi > ri {
		t.Errorf("want golang (higher usage count) rendered before rss, got:\n%s", body)
	}

	if !strings.Contains(body, `class="col-12 col-md-6"`) {
		t.Errorf("want a 2-column layout, got:\n%s", body)
	}
	if strings.Contains(body, "col-lg-4") {
		t.Errorf("want no 3-column breakpoint, got:\n%s", body)
	}
}

// TestHandleTagListViewRejectsPageZero verifies that page=0 is rejected
// instead of underflowing the unsigned offset arithmetic.
func TestHandleTagListViewRejectsPageZero(t *testing.T) {
	tags := []taxonomy.Tag{
		{UUID: "tag-1", UserUUID: testTagsCtxUser.UUID, Name: "golang"},
	}
	tc := newTestTagsController(tags, nil, nil)

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tags?page=0", nil)
	ctx := httpcontext.WithUser(r.Context(), testTagsCtxUser)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	tc.handleTagListView()(w, r)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Location"); got != "/tags" {
		t.Errorf("want redirect to /tags, got %q", got)
	}
}

// TestHandleTagListViewLimitsPageSize verifies that the tag list page caps
// at 60 tags per page.
func TestHandleTagListViewLimitsPageSize(t *testing.T) {
	fake := faker.New()

	const wantTagCount = 61
	tags := make([]taxonomy.Tag, wantTagCount)
	for i := range tags {
		tags[i] = taxonomy.Tag{
			UUID:     fake.UUID().V4(),
			UserUUID: testTagsCtxUser.UUID,
			Name:     fmt.Sprintf("tag-%02d", i),
		}
	}

	tc := newTestTagsController(tags, nil, nil)

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tags", nil)
	ctx := httpcontext.WithUser(r.Context(), testTagsCtxUser)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	tc.handleTagListView()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if got := strings.Count(body, "tag-row-"); got != 60 {
		t.Errorf("want 60 tag rows on the first page, got %d, body:\n%s", got, body)
	}
	if !strings.Contains(body, `aria-label="Goto page 2"`) {
		t.Errorf("want a second page to exist for the 61st tag, got:\n%s", body)
	}
}

// TestHandleTagListViewSortsByUsageCount verifies that tags are sorted by
// descending total usage (bookmark count + feed count), and that tags with
// equal usage counts fall back to alphabetical order.
func TestHandleTagListViewSortsByUsageCount(t *testing.T) {
	fake := faker.New()
	lowUUID := fake.UUID().V4()
	highUUID := fake.UUID().V4()
	tieAUUID := fake.UUID().V4()
	tieBUUID := fake.UUID().V4()

	tags := []taxonomy.Tag{
		{UUID: lowUUID, UserUUID: testTagsCtxUser.UUID, Name: "low"},
		{UUID: highUUID, UserUUID: testTagsCtxUser.UUID, Name: "high"},
		{UUID: tieAUUID, UserUUID: testTagsCtxUser.UUID, Name: "tie-a"},
		{UUID: tieBUUID, UserUUID: testTagsCtxUser.UUID, Name: "tie-b"},
	}
	// "low" is never attached to a bookmark or subscription, giving it a
	// usage count of 0 — unambiguously last regardless of name ordering.
	bookmarks := []bookmark.Bookmark{
		{UID: "b1", UserUUID: testTagsCtxUser.UUID, Title: "Bookmark 1", URL: "https://example1.tld", Tags: []string{"high", "tie-a"}},
		{UID: "b2", UserUUID: testTagsCtxUser.UUID, Title: "Bookmark 2", URL: "https://example2.tld", Tags: []string{"high", "tie-b"}},
	}

	tc := newTestTagsController(tags, bookmarks, nil)

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tags", nil)
	ctx := httpcontext.WithUser(r.Context(), testTagsCtxUser)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	tc.handleTagListView()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	wantOrder := []string{highUUID, tieAUUID, tieBUUID, lowUUID}

	sortIndices := make(map[string]int, len(wantOrder))
	for _, uuid := range wantOrder {
		idx := strings.Index(body, "tag-row-"+uuid)
		if idx < 0 {
			t.Fatalf("want tag row %q rendered, got:\n%s", uuid, body)
		}
		sortIndices[uuid] = idx
	}

	for i := 0; i < len(wantOrder)-1; i++ {
		if sortIndices[wantOrder[i]] > sortIndices[wantOrder[i+1]] {
			t.Errorf("want tag rows ordered %v by index, got indices %v", wantOrder, sortIndices)
		}
	}
}

func TestHandleTagEditView(t *testing.T) {
	fake := faker.New()

	t.Run("plain browser request renders the full page", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		r := newTagUUIDRequest(t, http.MethodGet, tagUUID, nil, false)
		w := httptest.NewRecorder()

		tc.handleTagEditView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, `value="golang"`) {
			t.Errorf("want the tag name pre-filled, got:\n%s", body)
		}
	})

	t.Run("htmx request renders only the form fragment", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		r := newTagUUIDRequest(t, http.MethodGet, tagUUID, nil, true)
		w := httptest.NewRecorder()

		tc.handleTagEditView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/tags/`+tagUUID+`/edit"`) {
			t.Errorf("want the modal fragment's form to be htmx-enhanced, got:\n%s", body)
		}
	})
}

func TestHandleTagEdit(t *testing.T) {
	fake := faker.New()

	t.Run("plain browser request renames and redirects", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		form := url.Values{"name": {"go"}}
		r := newTagUUIDRequest(t, http.MethodPost, tagUUID, form, false)
		w := httptest.NewRecorder()

		tc.handleTagEdit()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/tags" {
			t.Errorf("want redirect to /tags, got %q", got)
		}

		got, err := tc.taxonomyService.TagByUUID(t.Context(), testTagsCtxUser.UUID, tagUUID)
		if err != nil {
			t.Fatalf("failed to retrieve renamed tag: %q", err)
		}
		if got.Name != "go" {
			t.Errorf("want renamed tag name %q, got %q", "go", got.Name)
		}
	})

	t.Run("htmx request re-renders the tag row and closes the modal", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		form := url.Values{"name": {"go"}}
		r := newTagUUIDRequest(t, http.MethodPost, tagUUID, form, true)
		w := httptest.NewRecorder()

		tc.handleTagEdit()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Retarget"); got != "[id='tag-row-"+tagUUID+"']" {
			t.Errorf("want HX-Retarget targeting the tag row, got %q", got)
		}
		if got := w.Header().Get("HX-Reswap"); got != "outerHTML" {
			t.Errorf("want HX-Reswap outerHTML, got %q", got)
		}
		if got := w.Header().Get("HX-Trigger"); got != "modal:close" {
			t.Errorf("want HX-Trigger modal:close, got %q", got)
		}

		body := w.Body.String()
		if !strings.Contains(body, "go") {
			t.Errorf("want the renamed tag rendered, got:\n%s", body)
		}
	})

	t.Run("htmx request renaming into an existing tag re-renders the merged row", func(t *testing.T) {
		golangUUID := fake.UUID().V4()
		goUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{
			{UUID: golangUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"},
			{UUID: goUUID, UserUUID: testTagsCtxUser.UUID, Name: "go"},
		}
		tc := newTestTagsController(tags, nil, nil)

		form := url.Values{"name": {"go"}}
		r := newTagUUIDRequest(t, http.MethodPost, golangUUID, form, true)
		w := httptest.NewRecorder()

		tc.handleTagEdit()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		if got := w.Header().Get("HX-Retarget"); got != "[id='tag-row-"+golangUUID+"']" {
			t.Errorf("want HX-Retarget targeting the original tag row, got %q", got)
		}

		body := w.Body.String()
		if !strings.Contains(body, "id=\"tag-row-"+goUUID+"\"") {
			t.Errorf("want the merged-into tag's row rendered, got:\n%s", body)
		}
		if strings.Contains(body, "failed to retrieve") {
			t.Errorf("want no lookup-failure error after a merge, got:\n%s", body)
		}
	})
}

func TestHandleTagDeleteView(t *testing.T) {
	fake := faker.New()

	t.Run("plain browser request renders the full page", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		r := newTagUUIDRequest(t, http.MethodGet, tagUUID, nil, false)
		w := httptest.NewRecorder()

		tc.handleTagDeleteView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a full page (with layout), got:\n%s", body)
		}
		if !strings.Contains(body, "golang") {
			t.Errorf("want the tag name rendered, got:\n%s", body)
		}
	})

	t.Run("htmx request renders only the form fragment", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		r := newTagUUIDRequest(t, http.MethodGet, tagUUID, nil, true)
		w := httptest.NewRecorder()

		tc.handleTagDeleteView()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("want a fragment with no layout, got:\n%s", body)
		}
		if !strings.Contains(body, `hx-post="/tags/`+tagUUID+`/delete"`) {
			t.Errorf("want the modal fragment's form to be htmx-enhanced, got:\n%s", body)
		}
	})
}

func TestHandleTagDelete(t *testing.T) {
	fake := faker.New()

	t.Run("plain browser request deletes and redirects", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		r := newTagUUIDRequest(t, http.MethodPost, tagUUID, url.Values{}, false)
		w := httptest.NewRecorder()

		tc.handleTagDelete()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status 303, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Location"); got != "/tags" {
			t.Errorf("want redirect to /tags, got %q", got)
		}

		_, err := tc.taxonomyService.TagByUUID(t.Context(), testTagsCtxUser.UUID, tagUUID)
		if err == nil {
			t.Fatal("want the tag to be deleted, but it still exists")
		}
	})

	t.Run("htmx request removes the tag row and closes the modal", func(t *testing.T) {
		tagUUID := fake.UUID().V4()
		tags := []taxonomy.Tag{{UUID: tagUUID, UserUUID: testTagsCtxUser.UUID, Name: "golang"}}
		tc := newTestTagsController(tags, nil, nil)

		r := newTagUUIDRequest(t, http.MethodPost, tagUUID, url.Values{}, true)
		w := httptest.NewRecorder()

		tc.handleTagDelete()(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("HX-Retarget"); got != "[id='tag-row-"+tagUUID+"']" {
			t.Errorf("want HX-Retarget targeting the tag row, got %q", got)
		}
		if got := w.Header().Get("HX-Reswap"); got != "outerHTML" {
			t.Errorf("want HX-Reswap outerHTML, got %q", got)
		}
		if got := w.Header().Get("HX-Trigger"); got != "modal:close" {
			t.Errorf("want HX-Trigger modal:close, got %q", got)
		}

		_, err := tc.taxonomyService.TagByUUID(t.Context(), testTagsCtxUser.UUID, tagUUID)
		if err == nil {
			t.Fatal("want the tag to be deleted, but it still exists")
		}
	})
}
