#!/usr/bin/env bash

if [ ! -f "login-monitor" ]; then
  go build
fi

./login-monitor --config example-config.json --gmail-oauth2-config client-secret.json --gmail-oauth2-token token.json