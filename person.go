package githubpeople

import (
	"errors"
	"githubpeople/github"
	"strings"
)

type Person struct {
	MuenchenUser string `json:"muenchenUser"`
	GithubUser   string `json:"githubUser"`

	LdapBestätigt         bool `json:"ldapBestätigt"`
	GithubBestätigt       bool `json:"githubBestätigt"`
	GithubPeopleBestätigt bool `json:"githubPeopleBestätigt"`
}

func PersonFromGithubUser(g github.User) (p Person) {
	p.GithubUser = g.Login

	muechenUser, ok := strings.CutSuffix(g.Email, "@muenchen.de")
	if ok {
		p.MuenchenUser = muechenUser
	}

	return p
}

var ErrKeinMuenchenUser = errors.New("kein Benutzer in München angegeben")
var ErrKeinGithubUser = errors.New("kein Benutzer für Github angegeben")
var ErrNichtImLdapGefunden = errors.New("nicht im Ldap gefunden")
var ErrNichtInGithubPeopleGelistet = errors.New("nicht in GithubPeople gelistet")
var ErrNichtInGithubGefunden = errors.New("kein Benutzer in Github gefunden")

type ErrorGroup struct {
	Errors []error
}

func (g *ErrorGroup) Empty() bool {
	return len(g.Errors) == 0
}

func (g *ErrorGroup) Add(err error) {
	g.Errors = append(g.Errors, err)
}

func (g *ErrorGroup) Error() string {
	text := ""
	for i, err := range g.Errors {
		if i != 0 {
			text += ", "
		}
		text += err.Error()
	}
	return text
}

func (p Person) Validate() error {
	var g ErrorGroup

	if p.GithubPeopleBestätigt == false {
		g.Add(ErrNichtInGithubPeopleGelistet)
		// direkter fehler, wichtiger als die anderen
		return &g
	}

	if p.MuenchenUser == "" {
		g.Add(ErrKeinMuenchenUser)
	}
	if p.GithubUser == "" {
		g.Add(ErrKeinGithubUser)
	}

	if p.LdapBestätigt == false {
		g.Add(ErrNichtImLdapGefunden)
	}
	if p.GithubBestätigt == false {
		g.Add(ErrNichtInGithubGefunden)
	}

	if g.Empty() {
		return nil
	}
	return &g
}
