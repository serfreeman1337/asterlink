# AsterLink
Asterisk CRM Connector. 
Supports FreePBX v14 integration with [Bitrix24](https://github.com/serfreeman1337/asterlink/blob/master/README_bitrix24.md) and [SuiteCRM](https://github.com/serfreeman1337/asterlink/blob/master/README_suitecrm.md).

# Asterisk
You need Asterisk 13+.  
To monitor calls this connector listens for AMI events.

There should be 4 **different** contexts to distinguish calls:
* `incoming_context` - context for incoming calls from voip trunk. inbound calls will be registred there.
* `outgoing_context` - context for outgoing calls. outbound calls will be registred there.
* `ext_context` - extensions dials from queue, ring group, etc. use this context to route incoming calls to your extensions.
* `dial_context` - context for originating (click2dial) calls.

Default configuration is tested to work with FreePBX v14 and Asterisk v13.

Here is configuration for [basic-pbx](https://github.com/asterisk/asterisk/blob/master/configs/basic-pbx/extensions.conf) asterisk dialplan:
```yml
dialplan:
  incoming_context:
  - DCS-Incoming
  outgoing_context:
  - Outbound-Dial
  ext_context:
  - Dial-Users
  - DCS-Incoming
  dial_context: Long-Distance
```
You see `DCS-Incoming` in `ext_context` because we are dialing **queue** extensions directly from incoming context.  
(queue memeber config `member => PJSIP/1101` and not `member => Local/1101@Dial-Users` like in freepbx).

## CallerID Format
Connector can format CallerID using regexp. This useful when your VoIP provider doesn't send desired format. 

* `cid_format` - from PBX to CRM.
* `dial_format` - from CRM to PBX.
  * `expr` - regual expression (use double blackslashes).
  * `repl` - replace pattern.

If config is set and callerid doesn't matched any of regexp, then call will be ignored.

# CRM Integration
See instructions in the following files:
* [README_bitrix24.md](https://github.com/serfreeman1337/asterlink/blob/master/README_bitrix24.md) - For [Bitrix24](https://www.bitrix24.com/) Integration.
* [README_suitecrm.md](https://github.com/serfreeman1337/asterlink/blob/master/README_suitecrm.md) - For [SuiteCRM](https://suitecrm.com/) Integration.

# Install
Install asterlink under **/opt/asterlink** folder.
* Create folder /opt/asterlink:
  ```bash
  mkdir /opt/asterlink; cd /opt/asterlink
  ```
* Download binary from releases page:
  ```bash
  wget https://github.com/serfreeman1337/asterlink/releases/latest/download/asterlink_x86_64.tar.gz
  tar xvf asterlink_x86_64.tar.gz && rm asterlink_x86_64.tar.gz
  chmod +x asterlink
  ```
  * Or build it from source (assume you have [go installed](https://go.dev/doc/install)):
    ```bash
    go get github.com/serfreeman1337/asterlink
    go build github.com/serfreeman1337/asterlink
    ```
* Create configuration file. Use `conf.example.yml` as an example.
  ```bash
  wget https://raw.githubusercontent.com/serfreeman1337/asterlink/master/conf.example.yml
  mv conf.example.yml conf.yml
  nano conf.yml
  ```
  **Note:** config file is using YAML format and <ins>it requires to have proper indentation</ins>.  
  Use online yaml validator to check your file for errors.
* Test run:
  ```bash
  ./asterlink
  ```

## Startup script example
Create `/etc/systemd/system/asterlink.service` file with following contents:
```
[Unit]
Description=AsterLink Connector
After=freepbx.service

[Install]
WantedBy=multi-user.target

[Service]
ExecStart=/bin/sh -c 'exec /opt/asterlink/asterlink >>/opt/asterlink/app.log 2>>/opt/asterlink/err_app.log'
WorkingDirectory=/opt/asterlink
Restart=always
RestartSec=5
```
```bash
systemctl enable asterlink
systemctl start asterlink
```