package main

import (
	"fmt"
	"regexp"
	"testing"
)

const multi = `#If no profile is specified this one is used, in our case this means when running locally.
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/rome
    username: postgres
    password: postgres
    driverClassName: org.postgresql.Driver
    hikari:
      maximumPoolSize: 16
  liquibase:
    change-log: classpath:/db/changelog/changelog-master.yaml
    enabled: true
    drop-first: true
  jpa:
    show-sql: true
    hibernate:
      ddl-auto: none
      format_sql: true
    properties:
      hibernate:
        generate_statistics: true

  servlet:
    multipart:
      max-file-size: 50MB
      max-request-size: 50MB

logging:
  level:
    org:
      hibernate:
        SQL: 'error'
        SQL_SLOW: 'info'
        type: 'trace'

graphql:
  servlet:
    websocket:
      enabled: false
server:
  port: 8080

record-replay:
  enabled: false
  mode: REPLAY # RECORD/REPLAY/APPEND
  store: 'rome-bff/src/main/resources/crush-responses'

crush-service:
  mock: true
  mockResponseDelay: 1000
  url: http://crush.service

nemo-service:
  mock: true
  url: https://nemo.service
  url: https://nem-ox.dev.service.technipfmc.com
uns:
  mock: true
  createForSelf: true

access-management:
  mock: true

ad-service:
  mock: true

dgs:
  graphql:
    graphiql:
      title: 'Rome Graphiql'

vapid:
  publicKey: 'BK33QZT6Mexwvbxj81MxrWHqnmI4tKGT93FcyL4XACa_mnOOnj4rvoVFxjAsOR1uynDA3bKl6NdxRFW2ZlCtp08'
  privateKey: '-7lkaMDc8AODlU78HN8GW8p5GNVBqSYFnbQStH3wCIY'

mock-data:
  generate: true
  users:
    - uuid: 'a4132cb2-4967-3c50-9de9-e8fee19af64d'
      email: 'Killian.Wolf@localhost.com'
    - uuid: '394d0fe1-2520-4c37-9f5c-c101da79ecc5'
      email: 'johnny.oil@localhost.com'
    - uuid: '58051acb-70c0-4223-9f1c-974b79f4a122'
      email: 'Federico.Mendez@localhost.com'
    - uuid: '6ccac3f8-d5be-49da-b288-2602421411f4'
      email: 'Kiana.Arias@localhost.com'
    - uuid: '06cb8618-1d22-46ea-a51c-9ce69732e65f'
      email: 'Akaash.Owen@localhost.com'
    - uuid: '1ae55023-7375-493f-9031-490b8dc0d7ed'
      email: 'User@localhost.com'

email:
  mock: true

aws:
  s3:
    bucket: rome-filestack
    region: eu-west-1
    access-key:
    secret-key:

localstack:
  enabled: true
  s3:
    endpoint: http://s3.localhost.localstack.cloud:4566

miro:
  mock: true
  miroBaseUrl: https://miro.com
  miroApiBaseUrl: https://api.miro.com
  orgId:
  teamId:
  accessToken:
`

func TestRegexWorks(t *testing.T) {

	// ((?:https://|http://)([a-zA-Z0-9-]+)(?:.dev|.demo){0,1}.services)
	r := regexp.MustCompile("(?m)(?:http|https)://([a-zA-Z0-9-]+)(?:.dev|.demo){0,1}.service")
	m := r.FindAllStringSubmatch(multi, -1)
	fmt.Printf("%d matches\n", len(m))
	for _, s := range m {
		fmt.Println(s)
	}
}
