package githubpeople

import (
	"githubpeople/github"
)

type GithubPruefungService struct {
	GithubClient *github.Client
}

func (s *GithubPruefungService) GithubPruefung(people []Person) (gepruefte []Person, zusaetzlicheGefunden []Person, err error) {

	githubUsers, err := s.scrollGithubUsers()
	if err != nil {
		return nil, nil, err
	}
	byLogin := make(map[string]github.User)
	for _, u := range githubUsers {
		byLogin[u.Login] = u
	}

	gepruefte = make([]Person, len(people))
	for i, p := range people {
		_, ok := byLogin[p.GithubUser]
		if ok {
			delete(byLogin, p.GithubUser)
		}
		p.GithubBestätigt = ok
		gepruefte[i] = p
	}

	// zusätzlich gefundene benutzer
	for _, u := range byLogin {
		p := Person{
			GithubUser:      u.Login,
			GithubBestätigt: true,
		}
		zusaetzlicheGefunden = append(zusaetzlicheGefunden, p)
	}

	return gepruefte, zusaetzlicheGefunden, nil
}

func (s *GithubPruefungService) scrollGithubUsers() ([]github.User, error) {
	var (
		githubUsers = make([]github.User, 0)
		cursor      = github.CursorStart
	)
	for {
		next, lastCursor, err := s.GithubClient.ScrollOrganizationMembersGraphQL("it-at-m", cursor)
		if err != nil {
			return nil, err
		}
		cursor = lastCursor
		if len(next) == 0 {
			return githubUsers, nil
		}
		githubUsers = append(githubUsers, next...)
	}
}
