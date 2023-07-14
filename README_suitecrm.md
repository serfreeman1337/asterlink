# Asterisk SuiteCRM Integration
Features:
* Calls logging.
* Click-2-Call for phone fields.
* Pop up card for incoming and outgoing calls.
* Forwarding calls to assigned user.

##  Install
**For SuiteCRM 8 minimum working version is 8.2.0!**
* [Install and configure asterlink service](https://github.com/serfreeman1337/asterlink/blob/master/README.md) first.
* Uncomment `suitecrm` entry in `conf.yml` file and set:
  * `url` - SuiteCRM site address.
  * `endpoint_token` - any reasonable value.
  * `endpoint_addr` - listen address of the asterlink service.  
  **Note:** config file is using YAML format and it requires to have proper indentation.  
  Use online yaml validator to check your file for errors.
* Download [suitecrm-asterlink-module.zip](https://github.com/serfreeman1337/asterlink/releases/latest/download/suitecrm-asterlink-module.zip) archive from the [releases page](https://github.com/serfreeman1337/asterlink/releases).  
  **SuiteCRM 8:** Extension in the module archive was built with the `8.2.0` version and tested to work with the `8.3.1`.
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

**SuiteCRM 8:** If anything go wrong, simply delete `/extensions/asterlink` and `/cache` folders.

<details>
  <summary>
    SuiteCRM 8 extension build
  </summary>
  
  * Append following to `projects` entry in `angular.json` configuration:
    ```
    "asterlink": {
      "projectType": "application",
      "schematics": {
        "@schematics/angular:component": {
          "style": "css"
        },
        "@schematics/angular:application": {
          "strict": true
        }
      },
      "root": "extensions/asterlink/app",
      "sourceRoot": "extensions/asterlink/app/src",
      "prefix": "app",
      "architect": {
        "build": {
          "builder": "ngx-build-plus:browser",
          "options": {
            "outputPath": "extensions/asterlink/Resources/public",
            "index": "extensions/asterlink/app/src/index.html",
            "main": "extensions/asterlink/app/src/main.ts",
            "polyfills": "extensions/asterlink/app/src/polyfills.ts",
            "tsConfig": "extensions/asterlink/app/tsconfig.app.json",
            "inlineStyleLanguage": "css",
            "assets": [
              "extensions/asterlink/app/src/favicon.ico",
              "extensions/asterlink/app/src/assets"
            ],
            "styles": [
              "extensions/asterlink/app/src/styles.css"
            ],
            "scripts": [],
            "extraWebpackConfig": "extensions/asterlink/app/webpack.config.js",
            "commonChunk": false,
            "namedChunks": true,
            "sourceMap": true,
            "aot": true
          },
          "configurations": {
            "production": {
              "budgets": [
                {
                  "type": "initial",
                  "maximumWarning": "2mb",
                  "maximumError": "5mb"
                },
                {
                  "type": "anyComponentStyle",
                  "maximumWarning": "6kb",
                  "maximumError": "10kb"
                }
              ],
              "fileReplacements": [
                {
                  "replace": "extensions/asterlink/app/src/environments/environment.ts",
                  "with": "extensions/asterlink/app/src/environments/environment.prod.ts"
                }
              ],
              "outputHashing": "all",
              "extraWebpackConfig": "extensions/asterlink/app/webpack.prod.config.js",
              "optimization": true,
              "sourceMap": false,
              "namedChunks": true,
              "extractLicenses": true,
              "vendorChunk": false,
              "buildOptimizer": true
            },
            "dev": {
              "buildOptimizer": false,
              "optimization": false,
              "vendorChunk": false,
              "extractLicenses": false,
              "sourceMap": true,
              "outputPath": "public/extensions/asterlink"
            }
          },
          "defaultConfiguration": "production"
        },
        "serve": {
          "builder": "ngx-build-plus:dev-server",
          "configurations": {
            "production": {
              "browserTarget": "asterlink:build:production",
              "extraWebpackConfig": "extensions/asterlink/app/webpack.prod.config.js"
            },
            "development": {
              "browserTarget": "asterlink:build:development"
            }
          },
          "defaultConfiguration": "development",
          "options": {
            "port": 34000,
            "extraWebpackConfig": "extensions/asterlink/app/webpack.config.js"
          }
        },
        "extract-i18n": {
          "builder": "ngx-build-plus:extract-i18n",
          "options": {
            "browserTarget": "asterlink:build",
            "extraWebpackConfig": "extensions/asterlink/app/webpack.config.js"
          }
        },
        "test": {
          "builder": "@angular-devkit/build-angular:karma",
          "options": {
            "main": "extensions/asterlink/app/src/test.ts",
            "polyfills": "extensions/asterlink/app/src/polyfills.ts",
            "tsConfig": "extensions/asterlink/app/tsconfig.spec.json",
            "karmaConfig": "extensions/asterlink/app/karma.conf.js",
            "inlineStyleLanguage": "css",
            "assets": [
              "extensions/asterlink/app/src/favicon.ico",
              "extensions/asterlink/app/src/assets"
            ],
            "styles": [
              "extensions/asterlink/app/src/styles.css"
            ],
            "scripts": []
          }
        }
      }
    }
    ```
  * Append following to `scripts` in `package.json` configuration:
    ```
    "run:all": "node node_modules/@angular-architects/module-federation/src/server/mf-dev-server.js",
    "build-dev:asterlink": "ng build asterlink --configuration dev",
    "build:asterlink": "ng build asterlink --configuration production"
    ```
  * Follow [Front-end Developer Install Guide](https://docs.suitecrm.com/8.x/developer/installation-guide/front-end-installation-guide/).
  * Run `yarn run build:asterlink` to build extension.
  * More info: [Setting Up a Front-End Extension Module](https://docs.suitecrm.com/8.x/developer/extensions/frontend/fe-extensions-setup/)
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
