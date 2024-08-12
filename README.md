# githubpeople

[![Made with love by it@M][made-with-love-shield]][itm-opensource]

Compares members of a GitHub organization with users from LDAP and a map of GitHub and LDAP usernames.

## Built With

The project is built with technologies we use in our projects:

* Bash script
* [jq](https://github.com/jqlang/jq)
* ldapsearch from [ldap-utils](https://wiki.debian.org/LDAP/LDAPUtils)

## Roadmap

See the [open issues](https://github.com/it-at-m/githubpeople/issues) for a full list of proposed features (and known issues).

## Set up

Build docker image:

```sh
podman build . -t githubpeople
```

Run Docker container with environment variables and `githubpeople.json`:

```sh
podman run --env-file ./example/.env -v ./example/githubpeople.json:/app/githubpeople.json githubpeople
```

Optional it is possible to add a `.pem` certificate for LDAP:

```sh
podman run --env-file ./example/.env -v ./example/githubpeople.json:/app/githubpeople.json -v ./example/cert.pem:/app/cert.pem githubpeople
```

## Documentation

In the [it@M](https://github.com/it-at-m/) organization, members can develop with their own, private GitHub account. Therefore, the name of these accounts can be set individually and doesn't match the username in LDAP.

To keep track of the members in the organization, a list (`githubpeople.json`) was created that maps GitHub and LDAP usernames.

In order to check automatically if there are any differences between the organization members, the list and LDAP, this project was started.

It uses a Docker container, which runs a shell script that executes the following steps:

1. Verify that the member list (`githubpeople.json`) exists under the as an argument provided path

1. If a `.pem` certificate was provided, add it to `ldap.conf`

1. Fetch a list of organization members with GitHub's GraphQL API

1. Loop over the entries of `githubpeople.json`, checking if the user is in the organization and concatenating the LDAP usernames to a filter string

1. Search for users in LDAP with `ldapsearch` and the filter string so that only one request to LDAP is needed

1. Again, loop over `githubpeople.json` entries and check if the user is in LDAP

1. Loop over organization members and check if the member is on the `githubpeople.json` list

For each failing check, an error message is printed, and if at least one check fails, the script terminates with error code 1 instead of 0.

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please open an issue with the tag "enhancement", fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Open an issue with the tag "enhancement"
2. Fork the Project
3. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
4. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the Branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request

More about this in the [CODE_OF_CONDUCT](/CODE_OF_CONDUCT.md) file.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) file for more information.

## Contact

it@M - [opensource@muenchen.de](mailto:opensource@muenchen.de)

<!-- project shields / links -->
[made-with-love-shield]: https://img.shields.io/badge/made%20with%20%E2%9D%A4%20by-it%40M-yellow?style=for-the-badge
[itm-opensource]: https://opensource.muenchen.de/
