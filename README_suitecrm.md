# AsterLink SuiteCRM Integration
Features:
* Calls logging.
* Dialing for "tel" fields.
* Pop up card for incoming and outgoing calls.
* Forwarding calls to assigned user.

## Basic setup
Log calls for "Calls" module.

* Add new fields for Calls module
  * Administrator -> Studio -> Calls -> Fields
  * New fields:
    | Data Type | Field Name             | Display Label | Max Size |
    |-----------|------------------------|---------------|----------|
    | Integer   | asterlink_call_seconds | s.            | 2        |
    | TextField | asterlink_did          | DID           | 32       |
    | TextField | asterlink_uid          | Asterisk UID  | 64       |
    | Phone     | asterlink_cid          | CallerID      | 32       |
* Add new fields for Users module:
  * Administrator -> Studio -> Calls -> Fields
  * New fields:
    | Data Type | Field Name             | Display Label       | Max Size |
    |-----------|------------------------|---------------------|----------|
    | TextField | asterlink_ext          | Asterisk Extension  | 64       |
* Set ***customCode*** for ***duration_hours*** in **custom\modules\Calls\metadata\detailviewdefs.php**:
  ```php
  {$fields.duration_hours.value}{$MOD.LBL_HOURS_ABBREV} {$fields.duration_minutes.value}{$MOD.LBL_MINSS_ABBREV} {$fields.asterlink_call_seconds_c.value}{$MOD.LBL_ASTERLINK_CALL_SECONDS}&nbsp;
  ```
* Update layouts as you want.
* Set **asterlink_ext** in users profiles.
* Create new **Client Credentials Client** in Administrator -> OAuth2 Clients and Tokens
* Configure token ID and Secret in **config.yml**
* Do a test run. You should see userids from suitecrm in console.

## Click2Dial and Pop up card
* Configure **endpoint_addr** and **endpoint_token** in **config.yml**.
* Copy [connect/suitecrm/dist/asterlink](https://github.com/serfreeman1337/asterlink/tree/master/connect/suitecrm/dist) folder to ***suitecrm*** directory.
* Add following line to the end of the **custom/modules/logic_hooks.php** file:
  ```php
  $hook_array['after_ui_frame'][] = Array(1, 'asterlink javascript', 'asterlink/hooks.php', 'AsterLink', 'init_javascript');
  ```
* Add to **config_override.php**:
  ```php
  $sugar_config['asterlink']['endpoint_token'] = 'my_endpoint_token';
  $sugar_config['asterlink']['endpoint_url'] = 'http://my_endpoint_addr:my_endpoint_port';
  $sugar_config['asterlink']['endpoint_ws'] = 'ws://my_endpoint_addr:my_endpoint_port';
  ```
* Check "***Enable click-to-call for phone numbers Information***" in SuiteCRM System Settings.

### Apache2 endpoint proxy
* Enable mod_proxy, mod_proxy_http and mod_proxy_wstunnel.
* Config:
  ```
  ProxyPass	"/asterlink/ws/"	"ws://my_endpoint_addr:my_endpoint_port/ws/"
  ProxyPass	"/asterlink/"		"http://my_endpoint_addr:my_endpoint_port/"
  ```
* Update **config_override.php**:
  ```php
  $sugar_config['asterlink']['endpoint_url'] = 'http://apache_addr/asterlink';
  $sugar_config['asterlink']['endpoint_ws'] = 'ws://apache_addr/asterlink';
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
