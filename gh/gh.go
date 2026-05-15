package gh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Notification struct {
	ID        string    `json:"id"`
	Reason    string    `json:"reason"`
	Unread    bool      `json:"unread"`
	Title     string    `json:"title"`
	Repo      string    `json:"repo"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ghNotification struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Unread bool   `json:"unread"`
	Subject struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Type  string `json:"type"`
	} `json:"subject"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Client struct {
	token  string
	client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func (c *Client) GetNotifications(since time.Time) ([]Notification, error) {
	var all []ghNotification
	url := "https://api.github.com/notifications?all=false&participating=false&per_page=100"
	if !since.IsZero() {
		url += "&since=" + since.UTC().Format(time.RFC3339)
	}

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "github-notifier")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			resp.Body.Close()
			return nil, fmt.Errorf("invalid GitHub token")
		}
		if resp.StatusCode == 403 {
			resp.Body.Close()
			return nil, fmt.Errorf("rate limited")
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
		}

		var page []ghNotification
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		all = append(all, page...)

		url = ""
		link := resp.Header.Get("Link")
		if m := linkNextRe.FindStringSubmatch(link); len(m) > 1 {
			url = m[1]
		}
	}

	notifications := make([]Notification, 0, len(all))
	for _, n := range all {
		notifications = append(notifications, Notification{
			ID:        n.ID,
			Reason:    n.Reason,
			Unread:    n.Unread,
			Title:     n.Subject.Title,
			Repo:      n.Repository.FullName,
			URL:       apiURLToHTMLURL(n.Subject.URL, n.Subject.Type),
			UpdatedAt: n.UpdatedAt,
		})
	}

	return notifications, nil
}

func (c *Client) ValidateToken() error {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "github-notifier")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid GitHub token")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	return nil
}

func apiURLToHTMLURL(apiURL, subjectType string) string {
	const prefix = "https://api.github.com/repos/"
	if strings.HasPrefix(apiURL, prefix) {
		rest := apiURL[len(prefix):]
		if subjectType == "PullRequest" {
			rest = strings.Replace(rest, "/pulls/", "/pull/", 1)
		}
		return "https://github.com/" + rest
	}
	return apiURL
}
