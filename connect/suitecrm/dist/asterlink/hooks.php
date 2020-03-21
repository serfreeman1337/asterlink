<?php // serfreeman1337 // 21.03.2020
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

        if (!empty($_REQUEST['ajax_load'])) {
            if ($isDetail) {
                echo '<script>alInitFields()</script>';
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

    <script src="asterlink/script.js"></script>
    '.(($isDetail) ? '<script>alInitFields();</script>' : '').'
<!-- /AsterLink -->
';
    }
}