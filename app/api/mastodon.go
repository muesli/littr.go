package api

import (
	"encoding/json"
	"fmt"
	"github.com/mariusor/littr.go/app"
	"github.com/mariusor/littr.go/app/db"
	"github.com/mariusor/littr.go/app/models"
	"math"
	"net/http"
)

type stats struct {
	DomainCount int16 `json:"domain_count"`
	UserCount int16 `json:"user_count"`
	StatusCount int16 `json:"status_count"`
}

type desc struct {
	Description string `json:"description"`
	Email string `json:"email"`
	Stats stats `json:"stats"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Title string `json:"title"`
	Lang []string `json:"lang"`
	Uri string `json:"uri"`
	Urls []string `json:"urls,omitempty"`
	Version string `json:"version"`
}

// GET /api/v1/instance
// In order to be compatible with Mastodon
func ShowInstance(w http.ResponseWriter, r *http.Request) {
	ifErr := func (err... error) {
		if err != nil && len(err) > 0 && err[0] != nil {
			HandleError(w, r, http.StatusInternalServerError, err...)
			return
		}
	}

	u, err := db.Config.LoadAccounts(models.LoadAccountsFilter{
		MaxItems: math.MaxInt64,
	})
	ifErr(err)
	i, err := db.Config.LoadItems(models.LoadItemsFilter{
		MaxItems: math.MaxInt64,
	})
	ifErr(err)

	desc := desc {
		Title: "litter dot me",
		Description: "Littr.me is a link aggregator similar to reddit or hacker news",
		Email: "system@littr.me",
		Lang: []string{"en"},
		Stats: stats{
			DomainCount: 1,
			UserCount: int16(len(u)),
			StatusCount: int16(len(i)),
		},
		Uri: app.Instance.BaseURL,
		Version: fmt.Sprintf("2.5.0 compatible (littr.me %s)", app.Instance.Version),
	}
	data, err := json.Marshal(desc)
	ifErr(err)
	w.Header().Del("Cookie")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
