# AsterLink
Asterisk CRM Connector. 
Supports FreePBX v14 integration with Bitrix24 CRM.
## Asterisk
To monitor how call is going this connector listens for AMI events.
There should be 4 **different** contexts to distinguish calls:
* incoming_context
* outgoing_context
* ext_context - extensions dials from queue, ring group
* dial_context - for originate

Default configuration is tested to work with FreePBX v14 and Asterisk v13.
### CallerID Format
Connector can format CallerID using regexp. This useful when your VoIP provider doesn't send desired format. 

* cid_format - from PBX to CRM
* dial_format - from CRM to PBX
  * *expr* - regual expression (use double blackslashes)
  * *repl* - replace pattern

If config is set and callerid doesn't matched any of regexp, then call will be ignored.
## Bitrix24 Integration
### Basic setup
* Bitrix24 -> Applications -> Webhooks -> Add inbound webhook
  * Access permissions - check **crm**, **telephony**, **user**
  * Copy "**REST call example URL**" ***without*** "*profile/*" to **webhook_url** in config
* Bitrix24 -> Telephony -> Configure telephony -> Telephony users
  * Configure extensions for users
###  Call originations from Bitrix24
* Configure **webhook_endpoint_addr** in config
* Bitrix24 -> Applications -> Webhooks -> Add outbound webhook
  * Handler address - **http://my_endpoint_addr/originate/**
  * Event type - **ONEXTERNALCALLSTART**
  * Copy "**Authentication code**" to **webhook_originate_token** in config
* Bitrix24 -> Telephony -> Configure telephony -> Telephony settings
  * Default number for outgoing calls - Select your outbound hook
### Forwarding calls to assigned user
Dialplan example:
```
[bitrix24]
exten => route,1,Set(B24ASSIGNED=${CURL(http://my_endpoint_addr/assigned/${UNIQUEID})})
same => n,GotoIf($[${B24ASSIGNED}]?from-did-direct,${B24ASSIGNED},1)
same => n,Goto(ext-queues,400,1)
same => n,Hangup
```
### Recording upload
Bitrix24 can download and store call recording.
* Make your recording folder accessible. From webserver, for example:
  ```
  ln -s /var/spool/asterisk/monitor /var/www/html/recfiles
  ```
* Set **rec_upload** url in config
  ```
  http://my_pbx_addr/recfiles/
  ```