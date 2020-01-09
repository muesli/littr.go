package app

import (
	"context"
	"fmt"
	pub "github.com/go-ap/activitypub"
	"github.com/mariusor/littr.go/internal/log"
	"github.com/mariusor/qstring"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-ap/errors"
	"github.com/go-chi/chi"
)

const (
	MaxContentItems = 50
)

func isYay(v *Vote) bool {
	return v != nil && v.Weight > 0
}

func isNay(v *Vote) bool {
	return v != nil && v.Weight < 0
}

type aboutModel struct {
	Title string
	Desc  Desc
}

func parentLink(c Item) string {
	if c.Parent != nil {
		// @todo(marius) :link_generation:
		return fmt.Sprintf("/i/%s", c.Parent.Hash)
	}
	return ""
}

func opLink(c Item) string {
	if c.OP != nil {
		// @todo(marius) :link_generation:
		return fmt.Sprintf("/i/%s", c.OP.Hash)
	}
	return ""
}

func AccountLocalLink(a Account) string {
	handle := "anonymous"
	if len(a.Handle) > 0 {
		handle = a.Handle
	}
	// @todo(marius) :link_generation:
	return fmt.Sprintf("/~%s", handle)
}

// ShowAccountHandle
func ShowAccountHandle(a Account) string {
	//if strings.Contains(a.Handle, "@") {
	//	// @TODO(marius): simplify this at a higher level in the stack, see Account::FromActivityPub
	//	if parts := strings.SplitAfter(a.Handle, "@"); len(parts) > 1 {
	//		if strings.Contains(parts[1], app.Instance.HostName) {
	//			handle := parts[0]
	//			a.Handle = handle[:len(handle)-1]
	//		}
	//	}
	//}
	return a.Handle
}

// AccountPermaLink
func AccountPermaLink(a Account) string {
	if a.HasMetadata() && len(a.Metadata.URL) > 0 {
		return a.Metadata.URL
	}
	return AccountLocalLink(a)
}

// ItemPermaLink
func ItemPermaLink(i Item) string {
	if !i.IsLink() && i.HasMetadata() && len(i.Metadata.URL) > 0 {
		return i.Metadata.URL
	}
	return ItemLocalLink(i)
}

// ItemLocalLink
func ItemLocalLink(i Item) string {
	if i.SubmittedBy == nil {
		// @todo(marius) :link_generation:
		return fmt.Sprintf("/i/%s", i.Hash.Short())
	}
	return fmt.Sprintf("%s/%s", AccountLocalLink(*i.SubmittedBy), i.Hash.Short())
}

func followLink(f FollowRequest) string {
	return fmt.Sprintf("%s/%s", AccountLocalLink(*f.SubmittedBy), "follow")
}

// HandleIndex serves / request
func (h *handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	filter := Filters{
		LoadItemsFilter: LoadItemsFilter{
			InReplyTo: []string{""},
			Deleted:   []bool{false},
			Federated: []bool{false},
			Private:   []bool{false},
		},
		Page:     1,
		MaxItems: MaxContentItems,
	}
	if err := qstring.Unmarshal(r.URL.Query(), &filter); err != nil {
		h.logger.Debug("unable to load url parameters")
	}

	baseURL, _ := url.Parse(h.conf.BaseURL)
	title := fmt.Sprintf("%s: main page", baseURL.Host)

	acct := account(r)
	base := path.Base(r.URL.Path)
	switch strings.ToLower(base) {
	case "self":
		title = fmt.Sprintf("%s: self", baseURL.Host)
		h.logger.Debug("showing self posts")
	case "federated":
		title = fmt.Sprintf("%s: federated", baseURL.Host)
		h.logger.Debug("showing federated posts")
		filter.Federated = []bool{true}
	default:
	}
	m := itemListingModel{}
	m.Title = title
	m.HideText = true
	comments, err := loadItems(r.Context(), filter, acct, h.logger)
	if err != nil {
		h.v.HandleErrors(w, r, errors.NewNotValid(err, "Unable to load items!"))
	}
	for _, c := range comments {
		m.Items = append(m.Items, c)
	}
	if len(comments) >= filter.MaxItems {
		m.nextPage = filter.Page + 1
	}
	if filter.Page > 1 {
		m.prevPage = filter.Page - 1
	}
	h.v.RenderTemplate(r, w, "listing", m)
}

// HandleIndex serves / request
func (h *handler) HandleInbox(w http.ResponseWriter, r *http.Request) {
	filter := Filters{
		LoadItemsFilter: LoadItemsFilter{
			InReplyTo: []string{""},
			Deleted:   []bool{false},
			Federated: []bool{false},
			Private:   []bool{false},
		},
		Page:     1,
		MaxItems: MaxContentItems,
	}
	if err := qstring.Unmarshal(r.URL.Query(), &filter); err != nil {
		h.logger.Debug("unable to load url parameters")
	}

	baseURL, _ := url.Parse(h.conf.BaseURL)
	title := fmt.Sprintf("%s: main page", baseURL.Host)

	acct := account(r)
	title = fmt.Sprintf("%s: followed", baseURL.Host)

	filter.FollowedBy = acct.Hash.String()

	m := itemListingModel{}
	m.Title = title
	m.HideText = true

	requests, _, err := h.storage.LoadFollowRequests(acct, Filters{
		LoadFollowRequestsFilter: LoadFollowRequestsFilter{
			On: Hashes{Hash(acct.Metadata.ID)},
		},
	})
	if err != nil {
		h.v.HandleErrors(w, r, errors.NewNotValid(err, "Unable to load items!"))
	}
	for _, r := range requests {
		f := follow{r}
		m.Items = append(m.Items, &f)
	}
	comments, err := loadItems(r.Context(), filter, acct, h.logger)
	if err != nil {
		h.v.HandleErrors(w, r, errors.NewNotValid(err, "Unable to load items!"))
	}
	for _, c := range comments {
		m.Items = append(m.Items, c)
	}
	if len(comments) >= filter.MaxItems {
		m.nextPage = filter.Page + 1
	}
	if filter.Page > 1 {
		m.prevPage = filter.Page - 1
	}

	h.v.RenderTemplate(r, w, "listing", m)
}

func loadItems(c context.Context, filter Filters, acc *Account, l log.Logger) (comments, error) {
	repo, ok := ContextRepository(c)
	if !ok {
		err := errors.Errorf("could not load item repository from Context")
		return nil, err
	}
	contentItems, _, err := repo.LoadItems(filter)

	if err != nil {
		return nil, err
	}
	comments := loadComments(contentItems)
	if acc.IsLogged() {
		acc.Votes, _, err = repo.LoadVotes(Filters{
			LoadVotesFilter: LoadVotesFilter{
				AttributedTo: []Hash{acc.Hash},
				ItemKey:      comments.getItemsHashes(),
			},
			MaxItems: MaxContentItems,
		})
		if err != nil {
			l.Error(err.Error())
		}
	}
	return comments, nil
}

// HandleTags serves /tags/{tag} request
func (h *handler) HandleTags(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	filter := Filters{
		MaxItems: MaxContentItems,
		Page:     1,
	}
	acct := account(r)
	if len(tag) == 0 {
		h.v.HandleErrors(w, r, errors.BadRequestf("missing tag"))
	}
	filter.Content = "#" + tag
	filter.ContentMatchType = MatchFuzzy
	if err := qstring.Unmarshal(r.URL.Query(), &filter); err != nil {
		h.logger.Debug("unable to load url parameters")
	}
	baseURL, _ := url.Parse(h.conf.BaseURL)
	m := itemListingModel{}
	m.Title = fmt.Sprintf("%s: tagged as #%s", baseURL.Host, tag)
	comments, err := loadItems(r.Context(), filter, acct, h.logger)
	if err != nil {
		h.v.HandleErrors(w, r, errors.NewNotValid(err, "oops!"))
	}
	for _, c := range comments {
		m.Items = append(m.Items, c)
	}
	if len(comments) >= filter.MaxItems {
		m.nextPage = filter.Page + 1
	}
	if filter.Page > 1 {
		m.prevPage = filter.Page - 1
	}
	h.v.RenderTemplate(r, w, "listing", m)
}

// HandleDomains serves /domains/{domain} request
func (h *handler) HandleDomains(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")

	acct := account(r)
	filter := Filters{
		LoadItemsFilter: LoadItemsFilter{
			Context: []string{"0"},
		},
		MaxItems: MaxContentItems,
		Page:     1,
	}
	if len(domain) > 0 {
		filter.LoadItemsFilter.URL = domain
		filter.Type = pub.ActivityVocabularyTypes{pub.PageType}
	} else {
		filter.MediaType = []MimeType{MimeTypeMarkdown, MimeTypeText, MimeTypeHTML}
	}
	if err := qstring.Unmarshal(r.URL.Query(), &filter); err != nil {
		h.logger.Debug("unable to load url parameters")
	}
	baseURL, _ := url.Parse(h.conf.BaseURL)
	m := itemListingModel{}
	m.Title = fmt.Sprintf("%s: from %s", baseURL.Host, domain)
	m.HideText = true
	comments, err := loadItems(r.Context(), filter, acct, h.logger)
	if err != nil {
		h.v.HandleErrors(w, r, errors.NewNotValid(err, "oops!"))
	}
	for _, c := range comments {
		m.Items = append(m.Items, c)
	}
	if len(comments) >= filter.MaxItems {
		m.nextPage = filter.Page + 1
	}
	if filter.Page > 1 {
		m.prevPage = filter.Page - 1
	}
	h.v.RenderTemplate(r, w, "listing", m)
}
