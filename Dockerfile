FROM golang:1.21-alpine as build
RUN apk --update add tzdata
RUN apk --update add ca-certificates

# WORKDIR /usr/local/share/ca-certificates
# ADD http://ocsp.muenchen.de/pki/LHM-SUBCA2-v1.pem .
# ADD http://ocsp.muenchen.de/pki/LHM-SUBCA2-v2.pem .
# RUN update-ca-certificates

WORKDIR /app
COPY . .
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /githubpeople ./cli/githubpeople

FROM scratch
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

COPY --from=build /githubpeople .
# JSON Data must be mounted as volume 
CMD ["/githubpeople", "-people", "githubpeople.json"]