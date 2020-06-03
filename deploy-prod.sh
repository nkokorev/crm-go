#!/bin/sh
GOOS=linux GOARCH=amd64 go build -o /home/mex388/go/bin/crm-server main.go
ssh -p 25794 root@45.84.226.178 'systemctl stop crm-server'
scp -P 25794 /home/mex388/go/bin/crm-server root@45.84.226.178:/var/www/ratuscrm/server/

#ssh -p 25794 root@45.84.226.178 'rm -rf /var/www/ratuscrm/server/files/*'
#scp -P 25794 -r /home/mex388/go/src/github.com/nkokorev/crm-go/files/* root@45.84.226.178:/var/www/ratuscrm/server/files/

ssh -p 25794 root@45.84.226.178 'systemctl daemon-reload'
ssh -p 25794 root@45.84.226.178 'systemctl restart crm-server'

ssh -p 25794 root@45.84.226.178 'chown -R nginx:nginx /var/www/ratuscrm/server'
ssh -p 25794 root@45.84.226.178 'nginx -s reload'

echo "Start web: https://app.ratuscrm.com/"
echo "=== The end deploy crm-server ==="

#systemctl daemon-reload