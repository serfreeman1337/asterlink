<?php // serfreeman1337 // 15.06.21 //

use SuiteCRM\Search\SearchModules;

require_once('include/MVC/View/SugarView.php');

class ViewConfig extends SugarView
{
    /**
     * @see SugarView::_getModuleTitleParams()
     */
    protected function _getModuleTitleParams($browserTitle = false)
    {
        return [
           "<a href='index.php?module=Administration&action=index'>".translate('LBL_MODULE_NAME', 'Administration')."</a>",
           translate('LBL_ASTERLINK_CTITLE', 'Administration'),
        ];
    }
    
    /**
     * @see SugarView::preDisplay()
     */
    public function preDisplay()
    {
        global $current_user;
        
        if (!is_admin($current_user)
                && !is_admin_for_module($GLOBALS['current_user'], 'AsterLink')) {
            sugar_die("Unauthorized access to administration.");
        }
    }
    
    /**
     * @see SugarView::display()
     */
    public function display()
    {
        global $mod_strings;
        global $app_list_strings;
        global $app_strings;
        global $current_user;
        global $currentModule;
        global $sugar_config;
        
        echo $this->getModuleTitle(false);
        
        // $this->ss->assign("SITEURL", $sugar_config['site_url']);
        $this->ss->assign("MOD", $mod_strings);
        $this->ss->assign("APP", $app_strings);
        // $this->ss->assign("THEME", (string)SugarThemeRegistry::current());
        $this->ss->assign("RETURN_MODULE", "Administration");
        $this->ss->assign("RETURN_ACTION", "index");
        $this->ss->assign("MODULE", $currentModule);

        require_once('modules/ModuleBuilder/parsers/relationships/DeployedRelationships.php');
        $relationships = new DeployedRelationships('Calls');

        // Load relationships for Calls module

        $r = [];
        foreach($relationships->getRelationshipList() as $relationshipName) {
            $rel = $relationships->get($relationshipName)->getDefinition();

            $rel_data = [
                'module' => $rel['lhs_module'] == 'Calls' ? $rel['rhs_module'] : $rel['lhs_module'],
                'name' => $rel['lhs_table'] == 'calls' ? $rel['rhs_table'] : $rel['lhs_table'], // TODO: get real link name!!!
                'relationship_name' => $rel['relationship_name']
            ]; 

            $rel_data['title'] = $rel['rhs_key'] != 'parent_id' ?
                                    translate('LBL_MODULE_NAME', $rel_data['module']) :
                                    translate('LBL_LIST_RELATED_TO', 'Calls').' '.translate('LBL_MODULE_NAME', $rel_data['module']);
            

            if ($rel_data['module'] == 'Users')
                continue;

            $bean = BeanFactory::getBean($rel_data['module']);

            $rel_data['phone_fields'] = [];
            $rel_data['name_fields'] = [];
            
            // load fields from related module
            foreach ($bean->getFieldDefinitions() as $field_name => $field) {
                $field_type = 'phone_fields'; // look for phone fields

                if ($field['type'] != 'phone') { // look for name fields
                    if (in_array($field['type'], ['name', 'fullname', 'varchar'])) // look by field type 
                        $field_type = 'name_fields';
                    else
                        continue;
                }
                
                $rel_data[$field_type][$field_name] = rtrim(
                    translate($field['vname'] ?? '', $rel_data['module']),
                    ':' // remove ":" from the translation
                ).' ('.$field_name.')';
            }

            if (!empty($rel_data['phone_fields']) && !empty($rel_data['name_fields'])) {
                $r[] = $rel_data;
            }
        }

        $rel_modules = []; 
        foreach($r as $rel_data) {
            $rel_modules[$rel_data['relationship_name']] = [
                'id' => $rel_data['relationship_name'],
                'title' => $rel_data['title'],
                'name_fields' => $rel_data['name_fields'],
                'phone_fields' => $rel_data['phone_fields'],
            ];
        }

        $this->ss->assign('REL_MODULES', $rel_modules);
        
        // Am I Doing It Right ?
        if (isset($sugar_config['asterlink'])) {
            if (isset($sugar_config['asterlink']['endpoint_token']))
                $this->ss->assign("ENDPOINT_TOKEN", $sugar_config['asterlink']['endpoint_token']);
            
            if (isset($sugar_config['asterlink']['endpoint_url']))
                $this->ss->assign("ENDPOINT_URL", $sugar_config['asterlink']['endpoint_url']);

            if (isset($sugar_config['asterlink']['endpoint_ws']))
                $this->ss->assign("ENDPOINT_WS", $sugar_config['asterlink']['endpoint_ws']);
    
            if (isset($sugar_config['asterlink']['relate_once']) && $sugar_config['asterlink']['relate_once'])
                $this->ss->assign('RELATE_ONCE', true);

            if (isset($sugar_config['asterlink']['relationships']))
                $this->ss->assign('REL_CONFIG', $sugar_config['asterlink']['relationships']);
        }

        $this->ss->display("modules/AsterLink/tpls/config.tpl");
    }
}
