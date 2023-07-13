<?php // serfreeman1337 // 13.07.2023 //

if (!defined('sugarEntry') || !sugarEntry) die('Not A Valid Entry Point');

class AsterLink {
    public function __construct() {
        require_once('modules/AsterLink/utils.php');
    }

    public function init_javascript($event, $arguments) {
    	if (!empty($_REQUEST['to_pdf']) || !empty($_REQUEST['sugar_body_only']) || (!empty($_GET['module']) && $_GET['module'] == 'Emails')) {
            return;
        }

        $config = getConfig();

        if (!$config) {
            return;
        }

        $token = getUserToken($config['endpoint_token']);

        if(!$token) {
            return;
        }

        $isDetail = (!empty($_REQUEST['action']) && in_array($_REQUEST['action'], array('index', 'DetailView')));

        if (!empty($_REQUEST['ajax_load'])) { // SuiteCRM ajax loading.
            if ($isDetail) { // Re-enable click2dial.
                echo '<script>alInitFields();</script>';
            }

			return;
        }
        
 
        echo '
<!-- AsterLink -->
<script>';

        if (empty($config['proxy_enabled']) || !$config['proxy_enabled']) {
            echo '
    const ASTERLINK_TOKEN = \''.$token.'\';
    const ASTERLINK_URL = \''.htmlspecialchars($config['endpoint_url']).'/\';';
        } else {
            // What a mess.
            echo '
    const ASTERLINK_URL = location.protocol + \'//\' + 
            location.hostname + 
            (location.port ? \':\' + location.port : \'\' ) +
            location.pathname + \'?entryPoint=AsterLink&action=\';';
        }

        echo '
    const ASTERLINK_RELMODULES = {';
        // what have I done ...
        foreach ($config['relationships'] as $rel_config) {
            echo "
            '".$rel_config['module']."': {
                show_create: ".($rel_config['show_create'] ? 'true' : 'false').",
                phone_field: '".htmlspecialchars($rel_config['phone_fields'][0])."'
            },";
        }
        
        echo '
    };';
    
    echo '
</script>
<script src="'.getJSPath('modules/AsterLink/javascript/asterlink.js').'"></script>';
    if ($isDetail) {
        echo '
<script>alInitFields();</script>
    ';
    }

    echo '
<!-- /AsterLink -->
';
    }
}
