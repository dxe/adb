FUNCTIONS := hello-world community-reports
REGION := us-west-2

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
	sam deploy --parameter-overrides \
	  "SmtpUser=$$(aws ssm get-parameter --name smtp_user --with-decryption --region $(REGION) --query Parameter.Value --output text)" \
	  "SmtpPass=$$(aws ssm get-parameter --name smtp_pass --with-decryption --region $(REGION) --query Parameter.Value --output text)" \
	  "DbHost=$$(aws ssm get-parameter --name mysql_lambda_host --with-decryption --region $(REGION) --query Parameter.Value --output text)" \
	  "DbUser=$$(aws ssm get-parameter --name mysql_lambda_user --with-decryption --region $(REGION) --query Parameter.Value --output text)" \
	  "DbPassword=$$(aws ssm get-parameter --name mysql_lambda_password --with-decryption --region $(REGION) --query Parameter.Value --output text)"
