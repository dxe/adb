# Original: https://gist.github.com/soubrunorocha/ec30b7704d737a1797b0281e97967834

# Fail if any line fails and return an error so that building this Dockerfile
# will not swallow the script's errors.
set -e

#set the root password
DEFAULTPASS=""

#set some config to avoid prompting
sudo debconf-set-selections <<EOF
mysql-community-server mysql-community-server/root-pass password $DEFAULTPASS
mysql-community-server mysql-community-server/re-root-pass password $DEFAULTPASS
EOF

#get the mysql repository via wget
wget https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb

#set debian frontend to not prompt
export DEBIAN_FRONTEND=noninteractive

#config the package
sudo -E dpkg -i mysql-apt-config_0.8.29-1_all.deb

#update apt to get mysql repository
sudo apt update

#install mysql according to previous config
sudo -E apt install mysql-server mysql-client --assume-yes --force-yes

rm mysql-apt-config_0.8.29-1_all.deb
