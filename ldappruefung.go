package githubpeople

import (
	"githubpeople/ldap"
)

type LdapPruefungService struct {
	LdapConn *ldap.Conn
	BaseDN   string
}

func (s *LdapPruefungService) LdapPruefung(people []Person) (gepruefte []Person, err error) {

	ldapUsers, err := s.LdapConn.FindUsers(s.BaseDN, muenchenUsernames(people))
	if err != nil {
		return nil, err
	}
	byCN := make(map[string]ldap.User)
	for _, u := range ldapUsers {
		byCN[u.CN] = u
	}

	gepruefte = make([]Person, len(people))
	for i, p := range people {
		_, ok := byCN[p.MuenchenUser]
		p.LdapBestätigt = ok
		gepruefte[i] = p
	}
	return gepruefte, err
}

func muenchenUsernames(people []Person) (muenchenUsers []string) {
	muenchenUsers = make([]string, 0)
	for _, p := range people {
		if p.MuenchenUser == "" {
			continue
		}
		muenchenUsers = append(muenchenUsers, p.MuenchenUser)
	}
	return muenchenUsers
}
