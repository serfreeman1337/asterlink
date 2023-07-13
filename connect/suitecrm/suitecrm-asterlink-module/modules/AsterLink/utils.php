<?php // serfreeman1337 // 13.07.2023 //

if(!defined('sugarEntry') || !sugarEntry) die('Not A Valid Entry Point');

function getConfig() {
    global $sugar_config;

    if (empty($sugar_config['asterlink']) ||
        empty($sugar_config['asterlink']['endpoint_token']) ||
        empty($sugar_config['asterlink']['endpoint_url'])) {
        return null;
    }

    return $sugar_config['asterlink'];
}

function getUserToken($endpointToken) {
    global $current_user;

    if (empty($current_user->asterlink_ext_c)) {
        return null;
    }

    $base64url_encode = function($data) {
        return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
    };

    $header = $base64url_encode(json_encode([ 'typ' => 'JWT', 'alg' => 'HS256' ]));
    $payload = $base64url_encode(json_encode(['id' => $current_user->id ]));
    $signature = $base64url_encode(
        hash_hmac('sha256', $header.'.'.$payload, $endpointToken, true)
    );

    return ($header.'.'.$payload.'.'.$signature);
}