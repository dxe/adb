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
		go build -tags lambda.norpc -o bin/$@/bootstrap ./functions/$@
	cd bin/$@ && zip -j ../$@.zip bootstrap

clean:
	rm -rf bin .aws-sam

deploy: build
	sam deploy
