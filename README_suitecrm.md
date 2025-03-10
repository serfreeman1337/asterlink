# Asterisk SuiteCRM Integration
Features:
* Calls logging.
* Click-2-Call for phone fields.
* Pop up card for incoming and outgoing calls.
* Forwarding calls to assigned user.

##  Install
**For SuiteCRM 8 minimum working version is 8.8.0!**
* [Install and configure asterlink service](https://github.com/serfreeman1337/asterlink/blob/master/README.md) first.
* Uncomment `suitecrm` entry in `conf.yml` file and set:
  * `url` - SuiteCRM site address.
  * `endpoint_token` - any reasonable value.
  * `endpoint_addr` - listen address of the asterlink service.  
  **Note:** config file is using YAML format and it requires to have proper indentation.  
  Use online yaml validator to check your file for errors.
* Download [suitecrm-asterlink-module.zip](https://github.com/serfreeman1337/asterlink/releases/latest/download/suitecrm-asterlink-module.zip) archive from the [releases page](https://github.com/serfreeman1337/asterlink/releases).  
  **SuiteCRM 8:** Extension in the module archive was built and tested to work with the `8.8.0` version.
* Upload and install this module using **Module Loader** on the SuiteCRM Admin page.
* On the SuiteCRM Admin page open **AsterLink Connector** module settings and set:
  * `Token` - to `endpoint_token` value in the `conf.yml` file.
  * `Endpoint URL` - AsterLink service address (AsterLink uses http protocol, so it must start with `http://`).  
    **Note:** AsterLink service address must be reachable for users browsers or if you `Enable proxy`, for SuiteCRM site only.
* Optional: display seconds for call duration in detail view:  
  Click `Add seconds to call duration view field`.
* Set `Asterisk Extensions` for users (edit user profile).
* Do a test run of asterlink service. You should see userids from suitecrm in the console.  
  **Note:** You need to restart the asterlink app evertime you change asterisk extensions for users.

**SuiteCRM 8:** If for some reason suitecrm stop working after installing this module, delete `/extensions/asterlink` and `/cache` folders.

<details>
  <summary>
    SuiteCRM 8 extension build
  </summary>
  
  * Follow [Front-end Developer Install Guide](https://docs.suitecrm.com/8.x/developer/installation-guide/8.8.0-front-end-installation-guide/).
  * Run `yarn merge-angular-json`.
  * Run `yarn build:extension asterlink` to build this extension.
</details>

## Forwarding calls to assigned user
Dialplan example:
```
[assigned]
exten => route,1,Set(ASSIGNED=${CURL(http://my_endpoint_addr/assigned/${UNIQUEID})})
same => n,GotoIf($[${ASSIGNED}]?from-did-direct,${ASSIGNED},1)
same => n,Goto(ext-queues,400,1)
same => n,Hangup
```

## Upgrade from 0.4 version
* Just upload new module. SuiteCRM should handle upgrading.
* You might need to run "Quick Repair and Rebuild".

## Upgrade from 0.3 version
* Its highly recomended to backup DB before upgrading.
* Remove logic hook from **custom/modules/logic_hooks.php** by removing a line with the `asterlink javascript`.
* Remove any lines with `$sugar_config['asterlink']` from **config_override.php**.
* Delete **asterlink** folder from the suitecrm directory.
* Install AsterLink Module.
* Migrate relationships config from **conf.yml** to AsterLink Module settings.
