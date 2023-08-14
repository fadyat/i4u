#!/bin/bash


function i4u() {
  export "$(grep -v "^#" .env | xargs)"
  ./bin/i4u "$@"
}

i4u "$@"