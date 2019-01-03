echo "=== Initializing. This script assumes that git has been installed on this machine."

sudo true

echo "=== Fetching via wget and installing docker using install script"

wget -qO- https://get.docker.com/ | sh

echo "=== Installing docker-compose"

COMPOSE_VERSION=`git ls-remote https://github.com/docker/compose | grep refs/tags | grep -oP "[0-9]+\.[0-9][0-9]+\.[0-9]+$" | tail -n 1`
sudo sh -c "curl -L https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose"
sudo chmod +x /usr/local/bin/docker-compose
sudo sh -c "curl -L https://raw.githubusercontent.com/docker/compose/${COMPOSE_VERSION}/contrib/completion/bash/docker-compose > /etc/bash_completion.d/docker-compose"


sudo service docker start
sudo service docker restart

echo "=== Running docker compose up "

docker-compose up -d --build

echo "=== Done !"

exit 0