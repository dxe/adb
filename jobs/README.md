# jobs

AWS Lambda jobs for ADB, deployed with [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/).

## Prerequisites

- Go (preinstalled in the devcontainer)
- Python + pip (preinstalled in the devcontainer)
- AWS credentials with deploy permissions for `us-west-2`
- SSM parameters `smtp_user` and `smtp_pass` set in `us-west-2`
- SSM parameters `mysql_lambda_host`, `mysql_lambda_user`, and
  `mysql_lambda_password` set in `us-west-2` (used by `community-reports`).

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

## CI deploy permissions

The GitHub Actions `deploy-lambdas` job (see `.github/workflows/main.yml`) runs
`make deploy` as the `github-actions` IAM user. The permissions it needs to run
`sam deploy` against this stack were granted via the customer-managed policy
`adb-sam-deploy-jobs`, attached to that user. The policy was created by hand
(there is no IaC for the AWS account); to view or edit it, open the `us-east-1`
console (IAM is global): IAM → Policies → search `adb-sam-deploy-jobs`, or go to
<https://us-east-1.console.aws.amazon.com/iam/home#/policies/details/arn:aws:iam::521324062467:policy%2Fadb-sam-deploy-jobs>.

The commands that created and attached it:

```bash
# Create the policy from a local document (CloudFormation, Lambda, IAM roles,
# EventBridge, and EC2 describe — all scoped to the jobs stack's resources).
aws iam create-policy --policy-name adb-sam-deploy-jobs \
  --policy-document file://.../policy.json

# Attach it to the CI user.
aws iam attach-user-policy --user-name github-actions \
  --policy-arn arn:aws:iam::521324062467:policy/adb-sam-deploy-jobs
```
