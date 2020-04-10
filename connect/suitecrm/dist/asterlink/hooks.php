<?php // serfreeman1337 // 09.04.2020
if (!defined('sugarEntry') || !sugarEntry) die('Not A Valid Entry Point');

class AsterLink {
    function init_javascript($event, $arguments) {
        if (!empty($_REQUEST['to_pdf'])) {
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

        echo '
<!-- AsterLink -->
    <script>
        const ASTERLINK_TOKEN = "'.$sugar_config['asterlink']['endpoint_token'].'";
        const ASTERLINK_URL = "'.$sugar_config['asterlink']['endpoint_url'].'";
        const ASTERLINK_USER = "'.$current_user->id.'";
    </script>

    <script src="asterlink/c2d.js"></script>
    '.(($isDetail) ? '<script>alInitFields();</script>' : '').
    (($hasWs) ? '

    <script src="asterlink/ws.js"></script>
    <script>
        const ASTERLINK_WS = "'.$sugar_config['asterlink']['endpoint_ws'].'";
        alWs();
    </script>' : '') . '
<!-- /AsterLink -->
';
    }
}