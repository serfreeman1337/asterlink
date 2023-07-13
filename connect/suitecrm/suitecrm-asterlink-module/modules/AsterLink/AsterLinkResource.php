<?php // serfreeman1337 // 13.07.2023 //

if(!defined('sugarEntry') || !sugarEntry) die('Not A Valid Entry Point');

require_once('modules/AsterLink/utils.php');

$config = getConfig();

if (empty($_REQUEST['action']) || !$config) {
    http_response_code(400);
    die();
}

$token = getUserToken($config['endpoint_token']);

if (!$token) {
    http_response_code(204);
    die();
}

session_write_close(); // Prevent session lock by making it read-only.

switch ($_REQUEST['action']) {
    case 'originate': // Proxy originate request to asterlink.
        if ($_SERVER['REQUEST_METHOD'] != 'POST' || empty($_POST['phone'])) {
            http_response_code(400);
            die('Invalid request method or empty phone field');
        }

        $url = $config['endpoint_url'] . '/originate';
        $ch = curl_init();
        curl_setopt_array($ch, [
            CURLOPT_URL => $url,
            CURLOPT_POSTFIELDS => http_build_query([ 'phone' => $_POST['phone'] ]),
            CURLOPT_HTTPHEADER => [ 'X-AsterLink-Token: ' . $token ],
            CURLOPT_CONNECTTIMEOUT => 10
        ]);

        curl_exec($ch);
        $http_code = curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        curl_close($ch);

        if (!$http_code) {
            $http_code = 502;
            die('Unable to connect to the asterlink endpoint');
        }

        http_response_code($http_code);

        break;
    case 'stream': // Proxy call events stream.
        $url = $config['endpoint_url'] . '/stream?token=' . $token;

        ignore_user_abort(true); // Disable default disconnect checks.

        header('Content-Type: text/event-stream');
        header('Cache-Control: no-cache');
        header('X-Accel-Buffering: no');

        $ch = curl_init();
        curl_setopt_array($ch, [
            CURLOPT_URL => $url,
            CURLOPT_WRITEFUNCTION => function($ch, $data) {
                // Echo and then flush everything.
                echo $data;
                ob_end_flush();
                flush();
                return strlen($data);
            },
            CURLOPT_CONNECTTIMEOUT => 10
        ]);

        // Use multi curl for async request. 
        $mh = curl_multi_init();
        curl_multi_add_handle($mh, $ch);

        // Start stream proxy.
        do {
            // PHP needs to send something in order to detect client disconnect.
            echo "\r";
            ob_end_flush();
            flush();

            if (connection_aborted())
                break;

            $status = curl_multi_exec($mh, $active);
            if ($active) {
                curl_multi_select($mh, 5); // Wait 5 seconds and then check again if stream was closed or not.
            }
        } while ($active && $status == CURLM_OK);

        curl_multi_remove_handle($mh, $ch);
        curl_multi_close($mh);
        die();

        break;
    default: // SuiteCRM 8 frontend request.
        $response = [];

        if (empty($config['proxy_enabled']) || !$config['proxy_enabled']) {
            $response['token'] = getUserToken($config['endpoint_token']);
            $response['endpoint_url'] = $config['endpoint_url'];
        }

        $relations = [];

        foreach ($config['relationships'] as $rel_config) {
            $relations[$rel_config['module']] = [
                'show_create' => $rel_config['show_create'],
                'phone_field' => $rel_config['phone_fields'][0]
            ];
        }

        if (count($relations)) {
            $response['relations'] = $relations;
        }

        header('Content-Type: application/json');
        echo json_encode($response);
}

