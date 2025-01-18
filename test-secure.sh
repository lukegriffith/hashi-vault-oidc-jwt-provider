#!/bin/bash
curl -H "Authorization: Bearer $1" http://localhost:8080/secure | jq
