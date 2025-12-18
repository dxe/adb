# Server scripts directory

## Azure SSL CA

The azure-ssl-CA.pem file was created by following these instructions:
https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-connect-tls-ssl

This file can be passed to `-ssl-ca` flag of `mysql` CLI to verify the server's identity and keep the password and other
transmitted data safe.
