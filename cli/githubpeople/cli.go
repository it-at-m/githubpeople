package main

import (
	"githubpeople"
	"githubpeople/github"
	"githubpeople/ldap"
	"log"
	"net/url"
	"os"
)

type Config struct {
	// e.g. ldaps://example.de:636
	LdapUrl string
	// user/password for bind
	LdapUser     string
	LdapPassword string
	// e.g. DC=example,DC=de
	LdapBaseDN string

	GithubToken string
	// e.g. http://proxy.de:80
	GithubProxy string
}

type Cli struct {
	LdapConn     *ldap.Conn
	GithubClient *github.Client

	LdapPruefungService   githubpeople.LdapPruefungService
	GithubPruefungService githubpeople.GithubPruefungService
}

func ConfigFromEnv() (c Config) {
	c.LdapUrl = mustEnv("GITHUBPEOPLE_LDAP_URL")
	c.LdapUser = mustEnv("GITHUBPEOPLE_LDAP_USER")
	c.LdapPassword = mustEnv("GITHUBPEOPLE_LDAP_PASSWORD")
	c.LdapBaseDN = mustEnv("GITHUBPEOPLE_LDAP_BASEDN")

	c.GithubToken = mustEnv("GITHUBPEOPLE_GITHUB_TOKEN")
	c.GithubProxy = mustEnv("GITHUBPEOPLE_GITHUB_PROXY")

	return c
}

func mustEnv(env string) string {
	e, ok := os.LookupEnv(env)
	if !ok || e == "" {
		log.Fatalln("required env is not set:", env)
	}
	return e
}

func NewCli(config Config) *Cli {
	cli := &Cli{}

	ldapConn, err := ldap.Dial(config.LdapUrl, config.LdapUser, config.LdapPassword)
	if err != nil {
		log.Fatal(err)
	}
	cli.LdapConn = ldapConn
	cli.LdapPruefungService = githubpeople.LdapPruefungService{
		LdapConn: ldapConn,
		BaseDN:   config.LdapBaseDN,
	}

	if config.GithubProxy == "" {
		githubClient := github.NewClient(config.GithubToken)
		cli.GithubClient = githubClient
	} else {
		proxyUrl, err := url.Parse(config.GithubProxy)
		if err != nil {
			log.Fatal(err)
		}
		cli.GithubClient = github.NewClientWithProxy(config.GithubToken, proxyUrl)
	}
	cli.GithubPruefungService = githubpeople.GithubPruefungService{GithubClient: cli.GithubClient}

	return cli
}

func (cli *Cli) Close() {
	cli.LdapConn.Close()
}
