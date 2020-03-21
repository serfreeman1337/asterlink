## SuiteCRM Instructions

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
* Copy **asterlink** folder to ***suitecrm*** directory.
* Add following line to the end of the **custom/modules/logic_hooks.php** file:
  ```php
  $hook_array['after_ui_frame'][] = Array(1, 'asterlink javascript', 'asterlink/hooks.php', 'AsterLink', 'init_javascript');
  ```
* Add to **config_override.php**:
  ```php
  $sugar_config['asterlink']['endpoint_url'] = 'http://localhost:5678';
  $sugar_config['asterlink']['endpoint_token'] = 'test';
  ```
* Create new **Client Credentials Client** in Administrator -> OAuth2 Clients and Tokens
* Configure token ID and Secret in **config.yml**
* ...