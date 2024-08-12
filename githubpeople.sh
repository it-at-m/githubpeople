#!/bin/bash
fail=0

# Get memberlist from cmd args and verify that the file exists
memberlist=$1
if [ ! -f $1 ]; then
  echo "Liste in $memberlist nicht gefunden."
  exit $?
fi

maxiterations=$2
if [ ! $2 ]; then
  echo "Bitte maximale Anzahl an Iterationen angeben."
  exit 1
fi

# Find .pem certificate and load it for ldap
cert=$(find . -type f -name "*.pem" -print -quit)
if [ -f $cert ]; then
  echo "TLS_CACERT $cert" >> /etc/ldap/ldap.conf
  echo "Zertifikat $cert geladen."
else
  echo "Kein weiteres Zertifikat geladen."
fi

echo "Lade Daten von GitHub API..."

orgmembers=$(
  curl https://api.github.com/graphql -X POST -sS --fail-with-body \
  -H "Authorization: Bearer ${GITHUBPEOPLE_GITHUB_TOKEN}" \
  -H "Content-Type: application/json" \
  --proxy "${GITHUBPEOPLE_GITHUB_PROXY}" \
  -d "$(jq -c -n --arg query '
  {
    organization(login: "it-at-m") {
      membersWithRole(first: 100, after: "") {
        nodes {
          login
        }
      }
    }
  }' '{"query":$query}')"
)

if [ $? -gt 0 ]; then
  echo "Fehler beim Laden der Daten von GitHub API."
  exit 1
fi

ldapFilter=""

echo "Durchsuche Liste nach Nutzern, die nicht in der GitHub Organisation sind..."

i=0
while true ; do
  # Stop if all members in list are checked
  if jq .[$i] $memberlist | grep -q null; then
    break
  fi
  if [ $i -gt $maxiterations ]; then
    echo "Maximum von $maxiterations Iterationen wurde erreicht."
    exit 1
  fi

  # Get user data from memberlist
  muenchenUser=$(jq .[$i].muenchenUser $memberlist | tr -d \")
  githubUser=$(jq .[$i].githubUser $memberlist | tr -d \")

  ldapFilter="$ldapFilter(cn=$muenchenUser)" # Append user to filter

  # Check if githubUser is in organization
  if ! echo "$orgmembers" | grep $githubUser > /dev/null 2>&1; then
    echo " - Nicht in GitHub Organisation aber in Liste: $muenchenUser / $githubUser"
    fail=1
  fi

  # Go to next member in list
  (( i += 1 ))
done

echo "Lade Daten von LDAP..."

# Search ldap for all users in list
ldapResponse=$(ldapsearch -H ${GITHUBPEOPLE_LDAP_URL} -D ${GITHUBPEOPLE_LDAP_USER} -w ${GITHUBPEOPLE_LDAP_PASSWORD} -b ${GITHUBPEOPLE_LDAP_BASEDN} "(|$ldapFilter)" cn)

if [ $? -gt 0 ]; then
  echo "Fehler beim Laden der Daten von LDAP."
  exit 1
fi

echo "Durchsuche Liste nach Nutzern, die nicht in LDAP sind..."

i=0
while true ; do
  # Stop if all members in list are checked
  if jq .[$i] $memberlist | grep -q null; then
    break
  fi
  if [ $i -gt $maxiterations ]; then
    echo "Maximum von $maxiterations Iterationen wurde erreicht."
    exit 1
  fi

  muenchenUser=$(jq .[$i].muenchenUser $memberlist | tr -d \")
  githubUser=$(jq .[$i].githubUser $memberlist | tr -d \")

  # Check if user is in LDAP response
  if ! echo "$ldapResponse" | grep "cn: $muenchenUser" > /dev/null 2>&1; then
    echo " - Nicht in LDAP aber in Liste: $muenchenUser / $githubUser"
    fail=1
  fi

  # Go to next member in list
  (( i += 1 ))
done

echo "Durchsuche GitHub Organisation nach Nutern, die nicht in der Liste sind..."

i=0
while true ; do
  # Get current username from graphql response
  githubUser=$(echo "$orgmembers" | jq .data.organization.membersWithRole.nodes[$i].login | tr -d \")

  # Stop if all members of organization are checked
  if echo "$githubUser" | grep -q null; then
    break
  fi
  if [ $i -gt $maxiterations ]; then
    echo "Maximum von $maxiterations Iterationen wurde erreicht."
    exit 1
  fi

  # Check if user is in list
  if ! jq . $memberlist | grep -w $githubUser > /dev/null 2>&1; then
    echo " - Nicht in Liste aber in GitHub Organisation: $githubUser"
    fail=1
  fi

  # Go to next member in list
  (( i += 1 ))
done

if [ $fail -gt 0 ]; then
  echo "Unterschiede gefunden."
  exit 1
else
  echo "Keine Unterschiede gefunden."
  exit 0
fi