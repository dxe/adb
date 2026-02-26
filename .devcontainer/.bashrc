# Add git support for bash-completion.
source /usr/share/bash-completion/completions/git

### Useful aliases for development ###

# Run ADB CLI
alias adb="go run ./cli"
# Connect to local MySql database
alias db="mysql -u $DB_USER -h $DB_HOST -D $DB_NAME --password=$DB_PASSWORD"
