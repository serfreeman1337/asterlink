# AsterLink
Asterisk CRM Connector. 
Supports FreePBX v14 integration with Bitrix24 CRM.
## Asterisk
To monitor how call is going this connector listens for AMI events.
There should be 4 **different** contexts to distinguish calls:
* incoming_context
* outgoing_context
* ext_context - to call extension from direct, queue or ring group
* dial_context - for originate

Default configuration is tested to work with FreePBX v14 and Asterisk v13.
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