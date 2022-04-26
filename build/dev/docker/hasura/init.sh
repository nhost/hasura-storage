#!/bin/sh

 # we wait to make sure postgres is ready
 sleep 3

 graphql-engine serve &
