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

# SMTP and database secrets are no longer passed in at deploy time. The
# functions resolve them from SSM at runtime (see internal/secrets), keeping
# plaintext secrets off the deploy command line.
deploy: build
	sam deploy
