#!/bin/bash

BASE_URL=https://0gcp6wff56.execute-api.ap-southeast-2.amazonaws.com/Prod
NUM_REQ=500
CONCURRENT_REQ=10
IN=../in
OUT=../out

ab -p $IN/search-go.json -T 'application/json' -n 3 -c 1 $BASE_URL/go/warmup
ab -p $IN/search-python.json -T 'application/json' -n 3 -c 1 $BASE_URL/python/warmup
ab -p $IN/search-rust.json -T 'application/json' -n 3 -c 1 $BASE_URL/rust/warmup
ab -p $IN/search-node.json -T 'application/json' -n 3 -c 1 $BASE_URL/node/warmup

sleep 3

ab -p $IN/wave.json -T 'application/json' -e $OUT/go-wave.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/go/wave
sleep 3
ab -p $IN/wave.json -T 'application/json' -e $OUT/python-wave.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/python/wave
sleep 3
ab -p $IN/wave.json -T 'application/json' -e $OUT/rust-wave.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/rust/wave
sleep 3
ab -p $IN/wave.json -T 'application/json' -e $OUT/node-wave.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/node/wave

sleep 3

ab -p $IN/save-rust.json -T 'application/json' -e $OUT/rust-save.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/rust/save
sleep 3
ab -p $IN/save-node.json -T 'application/json' -e $OUT/node-save.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/node/save
sleep 3
ab -p $IN/save-python.json -T 'application/json' -e $OUT/python-save.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/python/save
sleep 3
ab -p $IN/save-go.json -T 'application/json' -e $OUT/go-save.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/go/save

sleep 3

ab -p $IN/search-python.json -T 'application/json' -e $OUT/python-search.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/python/search
sleep 3
ab -p $IN/search-go.json -T 'application/json' -e $OUT/go-search.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/go/search
sleep 3
ab -p $IN/search-node.json -T 'application/json' -e $OUT/node-search.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/node/search
sleep 3
ab -p $IN/search-rust.json -T 'application/json' -e $OUT/rust-search.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/rust/search

sleep 3

ab -p $IN/search-python.json -T 'application/json' -e $OUT/python-conc.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/python/conc
sleep 3
ab -p $IN/search-go.json -T 'application/json' -e $OUT/go-conc.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/go/conc
sleep 3
ab -p $IN/search-node.json -T 'application/json' -e $OUT/node-conc.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/node/conc
sleep 3
ab -p $IN/search-rust.json -T 'application/json' -e $OUT/rust-conc.csv -n $NUM_REQ -c $CONCURRENT_REQ $BASE_URL/rust/conc

date