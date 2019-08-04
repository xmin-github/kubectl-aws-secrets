#!/bin/bash

CREDENTIALS=$(aws sts assume-role  --role-arn arn:aws:iam::676682069537:role/AdminAccess  --profile stephen --role-session-name codebuild-kubectl --duration-seconds 900)
export AWS_ACCESS_KEY_ID="$(echo ${CREDENTIALS} | jq -r '.Credentials.AccessKeyId')"
echo  $AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY="$(echo ${CREDENTIALS} | jq -r '.Credentials.SecretAccessKey')"
echo  $AWS_SECRET_ACCESS_KEY
export AWS_SESSION_TOKEN="$(echo ${CREDENTIALS} | jq -r '.Credentials.SessionToken')"
echo $AWS_SESSION_TOKEN
export AWS_EXPIRATION=$(echo ${CREDENTIALS} | jq -r '.Credentials.Expiration')
