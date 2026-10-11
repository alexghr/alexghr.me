package web

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/alexghr/alexghr.me/internal/state"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Link        string    `xml:"link"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PublishDate string `xml:"pubDate,omitempty"`
}

const rssItemPage = 5

func (h handler) serverRSS(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.ListPosts(r.Context(), state.PostStatusPublished, rssItemPage, 0)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "error generating RSS feed", "error", err)
		h.renderError(w, r)
		return
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       "Alex Gherghisan",
			Link:        h.publicURL.String(),
			Description: "This is Alex's personal website",
		},
	}

	for _, post := range posts {
		link := h.publicURL.JoinPath("blog", post.Slug, "/")
		item := rssItem{
			Title:       post.Title,
			Link:        link.String(),
			GUID:        link.String(),
			Description: post.Description,
		}

		// just in case I forget to set published_at
		if !post.PublishedAt.IsZero() {
			item.PublishDate = post.PublishedAt.Format(time.RFC1123Z)
		}

		feed.Channel.Items = append(feed.Channel.Items, item)
	}

	data, err := xml.Marshal(feed)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "encode RSS", "error", err)
		h.renderError(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=1800")
	if _, err := w.Write(append([]byte(xml.Header), data...)); err != nil {
		h.logger.ErrorContext(r.Context(), "write RSS", "error", err)
	}
}
