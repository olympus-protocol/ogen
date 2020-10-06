#!/bin/bash

function configure_systemd() {

sudo rm -rf /etc/systemd/system/ogen.service

cat << EOF > /etc/systemd/system/ogen.service

[Unit]
Description=Ogen
After=network.target

[Service]
Type=simple
User=root
LimitNOFILE=1024
Restart=on-failure
RestartSec=10
ExecStart=/usr/local/bin/ogen --logfile --dashboard
WorkingDirectory=/root/.config/ogen
PermissionsStartOnly=true

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  sleep 3
}

arch=$(uname -i)

if [ "$arch" == 'x86_64' ];
then
export GO=go1.14.4.linux-amd64.tar.gz
fi

if [ "$arch" == 'aarch64' ];
then
export GO=go1.14.4.linux-arm64.tar.gz
fi

echo "Installing dependencies"

apt update &> /dev/null && apt install git build-essential -y &> /dev/null

if ! command -v go version &> /dev/null
then
    echo "Golang not installed, downloading..."
    cd /opt && curl https://storage.googleapis.com/golang/${GO} -o ${GO}
    tar zxf ${GO} && rm ${GO}
    ln -s /opt/go/bin/go /usr/bin/
    export GOPATH=/root/go
fi

echo "Downloading Ogen"

cd /opt || exit

rm -rf ogen

git clone https://github.com/olympus-protocol/ogen &> /dev/null

cd ogen && bash ./scripts/build.sh &> /dev/null && cp ogen /usr/local/bin

title="Ogen Installed"
instructions_first="The program is installed in the systemd services"
instructions_second="To start the program run 'service ogen start'"

mkdir -p /root/.config/ogen

rm -rf ogen

configure_systemd

printf %"$(tput cols)"s |tr " " "*"
printf %"$(tput cols)"s |tr " " " "
printf "%*s\n" $(((${#title}+$(tput cols))/2)) "$title"
printf "%*s\n" $(((${#instructions_first}+$(tput cols))/2)) "$instructions_first"
printf "%*s\n" $(((${#instructions_second}+$(tput cols))/2)) "$instructions_second"
printf %"$(tput cols)"s |tr " " " "
printf %"$(tput cols)"s |tr " " "*"