build:
	npm install ./layers/node/ngeohash
	sam build --beta-features

deploy: build
	if [ -f samconfig.toml ]; \
		then sam deploy; \
		else sam deploy -g; \
	fi

clean:
	rm -rf ./build
	rm -rf ./target

delete:
	sam delete
	rm samconfig.toml