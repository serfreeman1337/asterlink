<?php // serfreeman1337 // 15.06.21 //

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
        $hasWs = !empty($sugar_config['asterlink']['endpoint_ws']);

        if (!empty($_REQUEST['ajax_load'])) { // SuiteCRM ajax loading
            if ($isDetail) { // reenable click2dial
                echo '<script>alInitFields();</script>';
            }

            if ($hasWs && 
                !empty($_REQUEST['action']) && $_REQUEST['action'] == 'EditView'
            ) {
                echo '<script>alFillContact();</script>';
            }

			return;
        }

        $jwtHeader = json_encode(['typ' => 'JWT', 'alg' => 'HS256']);
        $jwtPayload = json_encode(['id' => $current_user->id]);

        $base64UrlHeader = str_replace(['+', '/', '='], ['-', '_', ''], base64_encode($jwtHeader));
        $base64UrlPayload = str_replace(['+', '/', '='], ['-', '_', ''], base64_encode($jwtPayload));

        $jwtSignature = hash_hmac('sha256', $base64UrlHeader.".".$base64UrlPayload, $sugar_config['asterlink']['endpoint_token'], true);
        $base64UrlSignature = str_replace(['+', '/', '='], ['-', '_', ''], base64_encode($jwtSignature));

        $jwt = $base64UrlHeader.".".$base64UrlPayload.".".$base64UrlSignature;
 
        echo '
<!-- AsterLink -->
    <script>
        const ASTERLINK_TOKEN = "'.$jwt.'";
        const ASTERLINK_URL = "'.$sugar_config['asterlink']['endpoint_url'].'";
    </script>

    <script src="'.getJSPath('modules/AsterLink/javascript/c2d.js').'"></script>
    '.(($isDetail) ? '<script>alInitFields();</script>' : '');

    if ($hasWs) {
        global $app_list_strings;


        echo '
            <script src="'.getJSPath('modules/AsterLink/javascript/ws.js').'"></script>
            <script>
                const ASTERLINK_WS = "'.$sugar_config['asterlink']['endpoint_ws'].'";
                const ASTERLINK_RELMODULES = {';
    
            // whate have I done ...
            foreach ($sugar_config['asterlink']['relationships'] as $rel_config) {
                $name = $app_list_strings['moduleListSingular'][$rel_config['module']] ?? 
                    ($app_list_strings['parent_type_display'][$rel_config['module']] ?? $rel_config['module']);
                echo "'".$rel_config['module']."': {
                    name: '".$name."',
                    show_create: ".($rel_config['show_create'] ? 'true' : 'false').",
                    phone_field: '".$rel_config['phone_fields'][0]."'
                },";
            }

        echo '};
                alWs();
            </script>
        ';
    }

    echo '<!-- /AsterLink -->
';
    }
}
