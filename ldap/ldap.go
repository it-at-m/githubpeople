package ldap

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

type Conn struct {
	conn *ldap.Conn
}

type User struct {
	CN string
}

func Dial(url string, username string, password string) (*Conn, error) {
	c, err := ldap.DialURL(url)
	if err != nil {
		return nil, err
	}

	err = c.Bind(username, password)
	if err != nil {
		c.Close()
		return nil, err
	}

	return &Conn{
		conn: c,
	}, nil
}

func (c *Conn) Close() {
	c.conn.Close()
}

var ErrUserNotFound = errors.New("user not found")

func (c *Conn) FindUser(baseDN string, uid string) (u User, err error) {

	then := time.Now()

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		"DC=muenchen,DC=de",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))", ldap.EscapeFilter(uid)),
		[]string{"cn"},
		nil,
	)

	r, err := c.conn.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(then.Sub(time.Now()))

	if len(r.Entries) != 1 {
		return u, ErrUserNotFound
	}

	u.CN = r.Entries[0].GetAttributeValue("cn")
	return u, nil
}

func (c *Conn) FindUsers(baseDN string, uids []string) (users []User, err error) {

	for len(uids) > 0 {
		var size = 100
		if len(uids) < size {
			size = len(uids)
		}
		var chunk []string
		chunk, uids = uids[:size], uids[size:]
		for i, uid := range chunk {
			chunk[i] = "(uid=" + ldap.EscapeFilter(uid) + ")"
		}
		query := fmt.Sprintf("(|%s)", strings.Join(chunk, ""))

		searchRequest := ldap.NewSearchRequest(
			"DC=muenchen,DC=de",
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(objectClass=organizationalPerson)(%s))", query),
			[]string{"cn"},
			nil,
		)

		r, err := c.conn.Search(searchRequest)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range r.Entries {
			cn := entry.GetAttributeValue("cn")
			users = append(users, User{CN: cn})
		}

	}

	return users, nil
}
