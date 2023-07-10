<?php // serfreeman1337 // 10.07.2023 //

if (!defined('sugarEntry') || !sugarEntry) die('Not A Valid Entry Point');

class AsterLink {
    function init_javascript($event, $arguments) {
    	if (!empty($_REQUEST['to_pdf']) || !empty($_REQUEST['sugar_body_only']) || (!empty($_GET['module']) && $_GET['module'] == 'Emails')) {
            return;
        }

        global $sugar_config;

        if (empty($sugar_config['asterlink']) ||
            empty($sugar_config['asterlink']['endpoint_token']) ||
            empty($sugar_config['asterlink']['endpoint_url'])
        ) {
            return;
        }

        global $current_user;

        if(!$current_user->asterlink_ext_c) {
            return;
        }

        $isDetail = (!empty($_REQUEST['action']) && in_array($_REQUEST['action'], array('index', 'DetailView')));

        if (!empty($_REQUEST['ajax_load'])) { // SuiteCRM ajax loading.
            if ($isDetail) { // Re-enable click2dial.
                echo '<script>alInitFields();</script>';
            }

			return;
        }

        $base64url_encode = function($data) {
            return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
        };

        $header = $base64url_encode(json_encode(['typ' => 'JWT', 'alg' => 'HS256']));
        $payload = $base64url_encode(json_encode(['id' => $current_user->id]));
        $signature = $base64url_encode(
            hash_hmac('sha256', $header.'.'.$payload, $sugar_config['asterlink']['endpoint_token'], true)
        );
 
        echo '
<!-- AsterLink -->
<script>
    const ASTERLINK_TOKEN = \''.($header.'.'.$payload.'.'.$signature).'\';
    const ASTERLINK_URL = \''.htmlspecialchars($sugar_config['asterlink']['endpoint_url']).'\';
    const ASTERLINK_WORKER = \'modules/AsterLink/javascript/asterlink.worker.js\';
    const ASTERLINK_STREAM = ASTERLINK_URL + \'/stream?token=\' + ASTERLINK_TOKEN;
    const ASTERLINK_RELMODULES = {';
        // what have I done ...
        foreach ($sugar_config['asterlink']['relationships'] as $rel_config) {
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
