package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	token  string
	client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		client: http.DefaultClient,
	}
}

func NewClientWithProxy(token string, proxyUrl *url.URL) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		},
	}
}

type User struct {
	Login string `json:"login"`
	Email string `json:"email"`
	Name  string `json:"name"`

	// nur optional bei graphql
	OrganizationVerifiedDomainEmails []string `json:"organizationVerifiedDomainEmails"`
}

type Page[T any] struct {
	Page     int
	Next     bool
	NextPage int
	Values   []T
}

func (c *Client) do(method string, path string, body io.Reader) (*http.Response, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   path,
	}

	r, err := http.NewRequest(method, u.String(), body)
	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	r.Header.Set("Accept", "application/vnd.github+json")

	if err != nil {
		return nil, err
	}
	return c.client.Do(r)
}

func (c *Client) ListOrganizationMembers(org string, page int) (p Page[User], err error) {
	path := fmt.Sprintf("(/orgs/%s/members", org)
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return p, err
	}
	if resp.StatusCode != http.StatusOK {
		return p, fmt.Errorf("failed to list organization members, http response status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&p.Values)
	if err != nil {
		return p, err
	}

	link := resp.Header.Get("Link")
	nextpage, err := parseNextPage(link)
	if err == errNextPageNotFound {
		p.Next = false
		return
	} else if err != nil {
		return p, err
	}

	p.Next = true
	p.NextPage = nextpage
	return p, nil
}

var errNextPageNotFound = errors.New("next page link not found in header")

func parseNextPage(linkHeader string) (nextpage int, err error) {
	for _, chunk := range strings.Split(linkHeader, ",") {
		u, rel, ok := strings.Cut(chunk, ";")
		if !ok {
			continue
		}
		u, rel = strings.Trim(u, "<> "), strings.TrimSpace(rel)

		if rel == "next" {
			nextLink, err := url.Parse(u)
			if err != nil {
				return 0, fmt.Errorf("bad next link header: %s", linkHeader)
			}

			pageText := nextLink.Query().Get("page")
			page, err := strconv.Atoi(pageText)

			if err != nil {
				return 0, fmt.Errorf("bad page attr in next link header: %s", linkHeader)
			}
			return page, nil
		}
	}

	return 0, errNextPageNotFound
}

type Cursor string

const CursorStart Cursor = ""

type GraphQLQuery[Variables any] struct {
	Query     string    `json:"query"`
	Variables Variables `json:"variables"`
}

func (q *GraphQLQuery[Variables]) JSONReader() (io.Reader, error) {
	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}

type GraphQLResponse[T any] struct {
	Data   T              `json:"data"`
	Errors []GraphQLError `json:"errors"`
}

type GraphQLError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e GraphQLError) Error() string {
	return fmt.Sprint(e.Type, e.Message)
}

const scrollOrganizationMembersQuery = `
query ($org: String!, $after: String!) {
	organization(login: $org) {
		membersWithRole(first: 100, after: $after) {
			edges {
				cursor
				node {
					login
					name
					email
					organizationVerifiedDomainEmails(login: $org)
				}
			}
		}
	}
}
`

func (c *Client) ScrollOrganizationMembersGraphQL(org string, after Cursor) (users []User, lastCursor Cursor, err error) {
	type params struct {
		Org   string `json:"org"`
		After string `json:"after"`
	}

	query := GraphQLQuery[params]{
		Query: scrollOrganizationMembersQuery,
		Variables: params{
			Org:   org,
			After: string(after),
		},
	}

	reader, err := query.JSONReader()
	if err != nil {
		return users, lastCursor, err
	}

	resp, err := c.do("POST", "/graphql", reader)
	if err != nil {
		return users, lastCursor, err
	}
	if resp.StatusCode != http.StatusOK {
		return users, lastCursor, fmt.Errorf("failed to scroll organization members, http response status: %s", resp.Status)
	}

	qlResponse := GraphQLResponse[struct {
		Organization struct {
			MembersWithRole struct {
				Edges []struct {
					Cursor Cursor `json:"cursor"`
					Node   struct {
						Login string `json:"login"`
						Name  string `json:"name"`
						Email string `json:"email"`
						// seems to work only for enterprise github users
						OrganizationVerifiedDomainEmails []string `json:"organizationVerifiedDomainEmails"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"membersWithRole"`
		} `json:"organization"`
	}]{}

	err = json.NewDecoder(resp.Body).Decode(&qlResponse)
	if err != nil {
		return users, lastCursor, err
	}

	if len(qlResponse.Errors) > 0 {
		var err error
		for _, e := range qlResponse.Errors {
			err = errors.Join(err, e)
		}
		return users, lastCursor, err
	}

	for _, edge := range qlResponse.Data.Organization.MembersWithRole.Edges {
		lastCursor = edge.Cursor
		u := User{
			Login: edge.Node.Login,
			Name:  edge.Node.Name,
			Email: edge.Node.Email,

			OrganizationVerifiedDomainEmails: edge.Node.OrganizationVerifiedDomainEmails,
		}
		users = append(users, u)
	}

	return users, lastCursor, nil
}
