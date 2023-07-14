# Asterisk Bitrix24 Integration

## Basic setup
* [Install and configure asterlink service](https://github.com/serfreeman1337/asterlink/blob/master/README.md) first.
* Uncomment `b24` entry in `conf.yml` file.
* Bitrix24 -> Applications -> Developer resources -> Other -> Add inbound webhook:
  * Access permissions - check **crm**, **telephony**, **user**.
  * Copy "**REST call example URL**" ***without*** "*profile/*" to **webhook_url** in config.
* Bitrix24 -> Telephony (open /telephony address) -> Configure telephony -> Telephony users:
  * Configure extensions for users.
## Call originations from Bitrix24
* Configure **webhook_endpoint_addr** in `conf.yml`.
* Bitrix24 -> Applications -> Webhooks -> Add outbound webhook
  * Handler address - **http://my_endpoint_addr/originate/**.
  * Event type - **ONEXTERNALCALLSTART**.
  * Copy "**Authentication code**" to **webhook_originate_token** in `conf.yml`.
* Bitrix24 -> Telephony -> Configure telephony -> Telephony settings.
  * Default number for outgoing calls - Select your outbound hook.
## Forwarding calls to assigned user
Dialplan example:
```
[bitrix24]
exten => route,1,Set(B24ASSIGNED=${CURL(http://my_endpoint_addr/assigned/${UNIQUEID})})
same => n,GotoIf($[${B24ASSIGNED}]?from-did-direct,${B24ASSIGNED},1)
same => n,Goto(ext-queues,400,1)
same => n,Hangup
```
## Recording upload
Bitrix24 can download and store call recording.
* Make your recording folder accessible. From webserver, for example:
  ```
  ln -s /var/spool/asterisk/monitor /var/www/html/recfiles
  ```
* Set **rec_upload** url in config:
  ```
  http://my_pbx_addr/recfiles/
  ```