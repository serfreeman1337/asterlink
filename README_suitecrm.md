# AsterLink SuiteCRM Integration
Features:
* Calls logging.
* Dialing for "tel" fields.
* Pop up card for incoming and outgoing calls.
* Forwarding calls to assigned user.

## Basic setup
* Download [suitecrm-asterlink-module.zip](https://github.com/serfreeman1337/asterlink/releases/latest/download/suitecrm-asterlink-module.zip) archive from the releases page.  
  Or create this archive by your own from the contents of the [connect/suitecrm/suitecrm-asterlink-module](https://github.com/serfreeman1337/asterlink/tree/master/connect/suitecrm/suitecrm-asterlink-module) folder.
* Upload and install this module on the **Module Loader** SuiteCRM Administrator page.
* Configure `Token` on the **AsterLink Settings** SuiteCRM Administrator page.  
  Token must be equal with the one in **conf.yml** file.
* Optional: display seconds for call duration in detail view:  
  Click `Add seconds to call duration view field` on the **AsterLink Settings** SuiteCRM Administrator page.
* Update layouts as you want.
* Set `Asterisk Extensions` in users profiles.
* Do a test run. You should see userids from suitecrm in console.

You need to restart asterlink app evertime you change asterisk extensions for users.

## Click2Dial and Pop up card
* Configure `endpoint_addr` in the  **conf.yml** file.
* Configure `Endpoint URL` for Click-to-Call function and `WebSocket URL` for pop up card on the **AsterLink Settings** SuiteCRM Administrator page.
* Check`Enable click-to-call for phone numbers Information` on the SuiteCRM System Settings page.  
  Note: SuiteCRM will only enable this for CallerID with begining plus sign. 

### Apache2 endpoint proxy
* Enable mod_proxy, mod_proxy_http and mod_proxy_wstunnel.
* Config:
  ```
  ProxyPass	"/asterlink/ws/"	"ws://my_endpoint_addr:my_endpoint_port/ws/"
  ProxyPass	"/asterlink/"		"http://my_endpoint_addr:my_endpoint_port/"
  ```
* Update AsterLink module settings:  
  ```
  Endpoint URL: http://my_endpoint_addr:my_endpoint_port/
  WebSocket URL: ws://my_endpoint_addr:my_endpoint_port/ws/
  ```

## Forwarding calls to assigned user
Dialplan example:
```
[assigned]
exten => route,1,Set(ASSIGNED=${CURL(http://my_endpoint_addr/assigned/${UNIQUEID})})
same => n,GotoIf($[${ASSIGNED}]?from-did-direct,${ASSIGNED},1)
same => n,Goto(ext-queues,400,1)
same => n,Hangup
```
