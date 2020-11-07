#!/bin/bash

headers=" --header 'Content-Type: application/json'"
signupCmd="curl localhost:9001/api/v1/oauth/signup-url --request POST --data '{\"account_name\":\"xxx\",\"given_name\":\"Ace\",\"family_name\":\"Base\"}' ${headers}"

printf "\n${signupCmd}\n"
eval $signupCmd
printf "\n"

loginCmd="curl localhost:9001/api/v1/oauth/login-url --request POST --data '{\"account_code\":\"xxx\"}' ${headers}"

printf "\n${loginCmd}\n"
eval $loginCmd
printf "\n"
