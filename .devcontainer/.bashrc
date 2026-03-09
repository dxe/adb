# Add git support for bash-completion. Assumes bash-completion apt package is
# installed. devcontainer.json should install it automatically.
source /usr/share/bash-completion/completions/git

### Useful aliases for development ###

# Run ADB CLI
alias adb="go run /workspace/cli"
# Connect to local MySql database
alias db="mysql -u $DB_USER -h $DB_HOST -D $DB_NAME --password=$DB_PASSWORD"
