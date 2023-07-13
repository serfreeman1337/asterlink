<?php // serfreeman1337 // 13.07.2023 //

class AsterLinkController extends SugarController
{

    public function pre_save()
    {
    }

    public function action_save()
    {
        global $sugar_config;
        global $current_user;
        
        if (!is_admin($current_user)
            && !is_admin_for_module($GLOBALS['current_user'], 'Emails')
            && !is_admin_for_module($GLOBALS['current_user'], 'Campaigns')) {
            sugar_die("Unauthorized access to administration.");
        }

        if (isset($_POST['modify_duration'])) {
            
            // create custom layout
            if (!file_exists('custom/modules/Calls/metadata/detailviewdefs.php')) {
                require_once 'modules/ModuleBuilder/parsers/ParserFactory.php';          
                $parser = ParserFactory::getParser(
                    'detailview', 
                    'Calls'
                );
                $parser->handleSave(false);
            }

            // replace customCode for duration_hours field
            $detailviewdefs = str_replace(
                '{$fields.duration_hours.value}{$MOD.LBL_HOURS_ABBREV} {$fields.duration_minutes.value}{$MOD.LBL_MINSS_ABBREV}&nbsp;',
                '{$fields.duration_hours.value}{$MOD.LBL_HOURS_ABBREV} {$fields.duration_minutes.value}{$MOD.LBL_MINSS_ABBREV} {$fields.asterlink_call_seconds_c.value}{$MOD.LBL_ASTERLINK_CALL_SECONDS}&nbsp;',
                file_get_contents('custom/modules/Calls/metadata/detailviewdefs.php')
            );

            file_put_contents('custom/modules/Calls/metadata/detailviewdefs.php', $detailviewdefs);

            // disable inline edit and requirement
            $verdefs_ext = '';

            if (!file_exists('custom/modules/Calls/Ext/Vardefs/vardefs.ext.php')) {
                $vardefs_ext = '<?php'.PHP_EOL;
            } else {
                $vardefs_ext = preg_replace('/^\$dictionary\[\'Call\'\]\[\'fields\'\]\[\'duration_hours\'\].*/m', '', 
                    file_get_contents('custom/modules/Calls/Ext/Vardefs/vardefs.ext.php')
                );
                $vardefs_ext = str_replace('?>', '', $vardefs_ext).PHP_EOL;
            }

            $vardefs_ext .= "\$dictionary['Call']['fields']['duration_hours']['required']=false;
\$dictionary['Call']['fields']['duration_hours']['inline_edit']=false;

?>";

            file_put_contents('custom/modules/Calls/Ext/Vardefs/vardefs.ext.php', $vardefs_ext);

            header("Location: index.php?module=AsterLink&action=config");
            return;
        }

        if (isset($_POST['endpoint_token'])) {
            $sugar_config['asterlink']['endpoint_token'] = trim($_POST['endpoint_token']);
        } else {
            unset($sugar_config['asterlink']['endpoint_token']);
        }

        if (isset($_POST['endpoint_url'])) {
            $url = trim($_POST['endpoint_url']);
            $url = rtrim($url, '/');
            $sugar_config['asterlink']['endpoint_url'] = $url;
        } else {
            unset($sugar_config['asterlink']['endpoint_url']);
        }

        if (isset($_POST['proxy_enabled'])) {
            $sugar_config['asterlink']['proxy_enabled'] = true;
        } else {
            $sugar_config['asterlink']['proxy_enabled'] = [];
        }

        if (isset($_POST['relate_once'])) {
            $sugar_config['asterlink']['relate_once'] = true;
        } else {
            $sugar_config['asterlink']['relate_once'] = [];
        }

        // set to empty array, will be removed if left empty
        $sugar_config['asterlink']['relationships'] = [];

        if (isset($_POST['rel'])) {
            require_once('modules/ModuleBuilder/parsers/relationships/DeployedRelationships.php');
            $relationships = new DeployedRelationships('Calls');
            
            foreach ($_POST['rel'] as $i => $relationshipName) {
                $rel = $relationships->get($relationshipName)->getDefinition();

                $sugar_config['asterlink']['relationships'][] = [
                    'rel_name' => $relationshipName,
                    'name' => (
                        ($rel['rhs_key'] != 'parent_id') ? 
                            (($rel['lhs_table'] == 'calls') ? $rel['rhs_table'] : $rel['lhs_table']) :
                            'calls'
                    ),
                    'module' => $rel['lhs_module'] == 'Calls' ? $rel['rhs_module'] : $rel['lhs_module'],
                    'is_parent' => ($rel['rhs_key'] == 'parent_id'),
                    'name_field' => $_POST['rel_name_field'][$i],
                    'phone_fields' => $_POST['rel_phone_fields'][$i],
                    'show_create' => isset($_POST['rel_show_create'][$i])
                ];
            }
        }

        rebuildConfigFile($sugar_config, $sugar_version);

        header("Location: index.php?module=AsterLink&action=config");
        return;
    }

    public function post_save()
    {
    }
}
