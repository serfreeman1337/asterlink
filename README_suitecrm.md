# AsterLink SuiteCRM Integration
Features:
* Calls logging.
* Dialing for "tel" fields.
* Pop up card for incoming and outgoing calls.
* Forwarding calls to assigned user.

## Basic setup
First of all, you need to setup V8 API in SuiteCRM. Follow [this](https://docs.suitecrm.com/developer/api/developer-setup-guide/json-api/#_before_you_start_calling_endpoints) guide (The "Before you start calling endpoints" part).

* There might be problems with contact search. See pull request [#8492](https://github.com/salesagility/SuiteCRM/pull/8492) for solution.
* You might get **404 API error**, if you are using SuiteCRM with **php-fcgi**. See pull request [#8486](https://github.com/salesagility/SuiteCRM/pull/8486) for solution.

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
  * Administrator -> Studio -> Users -> Fields
  * New fields:
    | Data Type | Field Name             | Display Label       | Max Size |
    |-----------|------------------------|---------------------|----------|
    | TextField | asterlink_ext          | Asterisk Extension  | 64       |
* Optional: display seconds for call duration in detail view.  
  Set `customCode` for `duration_hours` in **custom\modules\Calls\metadata\detailviewdefs.php**:
  ```php
  {$fields.duration_hours.value}{$MOD.LBL_HOURS_ABBREV} {$fields.duration_minutes.value}{$MOD.LBL_MINSS_ABBREV} {$fields.asterlink_call_seconds_c.value}{$MOD.LBL_ASTERLINK_CALL_SECONDS}&nbsp;
  ```
  If this file doesn't exist, click **save & deply** on Calls DetailView layout in studio.
* Update layouts as you want.
* Set `asterlink_ext` in users profiles.
* Create new **Client Credentials Client** in Administrator -> OAuth2 Clients and Tokens
* Configure token ID and Secret in **config.yml**
* Do a test run. You should see userids from suitecrm in console.

You need to restart asterlink app evertime you change `asterisk_ext` for users.

### Call record relationships
See `relationships` key in config.
```yml
  relationships:
  -
    module: Contacts
    module_name: Contact
    primary_module: false
    show_create: true
    name_field: full_name
    phone_fields: [phone_mobile, phone_work]
  -
    module: Prospects
    module_name: Target
    primary_module: true
    show_create: false
    name_field: full_name
    phone_fields: [phone_mobile, phone_work]
  relate_once: false
```
* `relationships` - modules to search for records in. Define every new module as shown in config example
  * `module` - SuiteCRM Module ID
  * `module_name` - module name, that will show on popup card
  * `primary_module` - set `true`, if this module is primary in relationship to "Calls" module
  * `show_create` - shows create new record link on popup card
  * `name_field` - module field for found record name
  * `phone_fields` - module fields to search for phone number
* `relate_once` - set `false` to search and relate records in all modules from `relationships`

## Click2Dial and Pop up card
* Configure `endpoint_addr` and `endpoint_token` in **config.yml**. `endpoint_token` is just an any secret word you want.  
  It must be equal in both **conf.yml** and **config_override.php**.
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
  Note: SuiteCRM will only enable this for CallerID with begining plus sign. 

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
