FUNCTIONS := hello-world
REGION := us-west-2

.PHONY: build clean deploy deps $(FUNCTIONS)

deps:
	pip install -r requirements.txt

build: $(FUNCTIONS)

$(FUNCTIONS):
	mkdir -p bin/$@
	GOWORK=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -tags lambda.norpc -o bin/$@/bootstrap ./functions/$@
	cd bin/$@ && zip -j ../$@.zip bootstrap

clean:
	rm -rf bin .aws-sam

deploy: build
	sam deploy --parameter-overrides \
	  "SmtpUser=$$(aws ssm get-parameter --name smtp_user --with-decryption --region $(REGION) --query Parameter.Value --output text)" \
	  "SmtpPass=$$(aws ssm get-parameter --name smtp_pass --with-decryption --region $(REGION) --query Parameter.Value --output text)"
