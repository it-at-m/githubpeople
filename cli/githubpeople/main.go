package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"githubpeople"
	"log"
	"os"
)

func main() {
	cli := NewCli(ConfigFromEnv())
	defer cli.Close()

	var peopleFile string
	flag.StringVar(&peopleFile, "people", "githubpeople.json", "json file with registered githubpeople n")
	flag.Parse()

	people, err := loadGithubPeople(peopleFile)
	if err != nil {
		log.Fatalln(err)
	}

	people, zusaetzlicheGithub, err := cli.GithubPruefungService.GithubPruefung(people)
	if err != nil {
		log.Fatal(err)
	}
	people = append(people, zusaetzlicheGithub...)

	people, err = cli.LdapPruefungService.LdapPruefung(people)
	if err != nil {
		log.Fatalln(err)
	}

	var (
		errLogger = log.New(os.Stderr, "", 0)
		errCount  = 0
	)
	for _, p := range people {

		err := p.Validate()
		if err != nil {
			errLogger.Println("-", p.GithubUser+":", err)
			errCount += 1
		}
	}

	if errCount != 0 {
		errLogger.Fatalf("%d Validierungsfehler bei %d Personen.", errCount, len(people))
	} else {
		fmt.Printf("%d Personen ohne Fehler validiert.", len(people))
	}
}

func loadGithubPeople(filepath string) (people []githubpeople.Person, err error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open github peoples %s: %w", filepath, err)
	}

	err = json.NewDecoder(file).Decode(&people)
	if err != nil {
		return nil, fmt.Errorf("github peoples file %s could not be parsed as json: %w", filepath, err)
	}

	for i := range people {
		people[i].GithubPeopleBestätigt = true
		people[i].LdapBestätigt = false
		people[i].GithubBestätigt = false
	}

	return people, nil
}
