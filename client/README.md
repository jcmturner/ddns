# Client Configurations

## Mikrotik
### Set up script
Update the configuration values in the script for the following:
```
### Configuration variables ###

:global ddnsapi "https://1v9rrx0784.execute-api.eu-west-2.amazonaws.com/v1/ddns/"
:global ddnsuser "USERNAME"
:global ddnspasswd "PASSWORD"

:global ddnsrecord "www"
:global ddnsdomain "jtnet.co.uk."

:global extint "pppoe-out1"

### ### ### ### ### ### ### ###
```
Run the following at the Mikrotik command line:
```
/system script add name=DDNS owner=admin policy=read,write,policy,test source={
PASTE CONTENT HERE
}
```
### Set up schedule
Run the following at the Mikrotik command line:
```
/system scheduler add interval=1m name=DDNS on-event="/system script run DDNS" policy=\
    ftp,reboot,read,write,policy,test,password,sniff,sensitive start-time=startup
```