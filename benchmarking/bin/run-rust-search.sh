#!/bin/bash

BASE_URL=https://xhxq6jaix9.execute-api.ap-southeast-2.amazonaws.com/Prod
NUM_REQ=500
CONCURRENT_REQ=10
IN=../in
OUT=../out

ab -p $IN/search-rust.json -T 'application/json' -n 3 -c 1 $BASE_URL/rust/warmup

sleep 3

ab -p $IN/search-rust.json -T 'application/json' -e $OUT/rust-search.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/rust/search

sleep 3

ab -p $IN/search-rust.json -T 'application/json' -e $OUT/rust-conc.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/rust/conc

date