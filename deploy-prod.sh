#!/bin/sh
#GOOS=linux GOARCH=amd64 go build
GOOS=linux GOARCH=amd64 go build -o /home/mex388/go/bin/crm-server main.go
ssh -p 25794 root@45.84.226.178 'systemctl stop crm-server'
scp -P 25794 /home/mex388/go/bin/crm-server root@45.84.226.178:/var/www/ratuscrm/server/
#scp -P 25794 /home/mex388/go/src/github.com/nkokorev/go-test/.env root@193.200.74.37:/var/www/go
#ssh -p 25794 root@193.200.74.37 'rm -rf /var/www/go/html'
ssh -p 25794 root@45.84.226.178 'systemctl daemon-reload'
ssh -p 25794 root@45.84.226.178 'systemctl restart crm-server'

ssh -p 25794 root@45.84.226.178 'chown -R nginx:nginx /var/www/ratuscrm/server'
ssh -p 25794 root@45.84.226.178 'nginx -s reload'

echo "Start web: http://app.ratuscrm.com/"
echo "=== The end deploy crm-server ==="

#systemctl daemon-reload