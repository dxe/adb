# jobs

AWS Lambda jobs for ADB, deployed with [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/).

## Prerequisites

- Go (preinstalled in the devcontainer)
- Python + pip (preinstalled in the devcontainer)
- AWS credentials with deploy permissions for `us-west-2`
- SSM parameters `smtp_user` and `smtp_pass` set in `us-west-2`
- SSM parameters `mysql_lambda_host`, `mysql_lambda_user`, and
  `mysql_lambda_password` set in `us-west-2` (used by `community-reports`).

The functions resolve these SecureString parameters from SSM at runtime (see
`internal/secrets`); they are not passed in at deploy time. The template grants
each function `ssm:GetParameter` + `kms:Decrypt` for only the parameters it uses.

Install Python dependencies (currently just `aws-sam-cli`):

```bash
cd jobs
make deps
```

Log in with the AWS CLI:

```bash
aws login
```

Note: the above command works for IAM users, not SSO / Identity Center users,
and uses your console credentials rather than storing a long-term IAM access
key. You must re-run it every time your console session expires.

## Functions

- `hello-world` — sends a "Hello World" email.
- `community-reports` — sends email reports to community team.

## Build

Cross-compiles each function to a `bootstrap` binary and zips it into `bin/`:

```bash
cd jobs
make build
```

## Deploy

```bash
cd jobs
make deploy
```

## Invoke (smoke test)

```bash
aws lambda invoke --function-name adb-jobs-hello-world --region us-west-2 /dev/stdout
aws lambda invoke --function-name adb-jobs-community-reports --region us-west-2 /dev/stdout
```

## Clean

```bash
cd jobs
make clean
```
