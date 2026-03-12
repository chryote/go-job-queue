#!/bin/bash

# $1 is the ID, $2 is the payload (e.g., "fail@example.com")
curl -X POST http://localhost:8080/enqueue \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$1\",
    \"type\": \"EMAIL\",
    \"payload\": \"$2\"
  }"
