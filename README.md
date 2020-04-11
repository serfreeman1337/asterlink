# AsterLink
Asterisk CRM Connector. 
Supports FreePBX v14 integration with [Bitrix24](https://github.com/serfreeman1337/asterlink/blob/master/README_bitrix24.md) and [SuiteCRM](https://github.com/serfreeman1337/asterlink/blob/master/README_suitecrm.md).

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