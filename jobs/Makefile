FUNCTIONS := hello-world community-reports

.PHONY: build clean deploy deps $(FUNCTIONS)

deps:
	pip install -r requirements.txt

build: $(FUNCTIONS)

$(FUNCTIONS):
	mkdir -p bin/$@
	# -trimpath and -buildvcs=false strip absolute paths and VCS stamps so the
	# binary is byte-identical when the source is unchanged, avoiding a new Lambda version on every deploy.
	GOWORK=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -trimpath -buildvcs=false -tags lambda.norpc -o bin/$@/bootstrap ./functions/$@
	# Pin the mtime and strip extra attributes (-X) so the zip is byte-identical
	# when the binary is unchanged, matching the reproducible build above and
	# avoiding a new Lambda version on every deploy.
	touch -t 198001010000 bin/$@/bootstrap
	rm -f bin/$@.zip && cd bin/$@ && zip -j -X ../$@.zip bootstrap

clean:
	rm -rf bin .aws-sam

deploy: build
	sam deploy
