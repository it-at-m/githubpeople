FROM debian:12

RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install curl -y
RUN apt-get install jq -y
RUN apt-get install ldap-utils -y

WORKDIR /app
COPY ./githubpeople.sh .

RUN chmod +x ./githubpeople.sh
CMD ["./githubpeople.sh", "githubpeople.json", "200"]