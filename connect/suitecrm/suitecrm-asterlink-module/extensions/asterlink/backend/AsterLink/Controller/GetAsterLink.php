<?php

namespace App\Extension\asterlink\backend\AsterLink\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\Security\Core\Security;

use Exception;

use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;


/*
 * 
 */
class GetAsterLink extends AbstractController {
    public function __invoke(Security $security): Response|array
    {
        global $sugar_config, $current_user;

        // AsterLink isn't configured or user doesn't have extension.
        if (empty($sugar_config['asterlink']) || 
            empty($sugar_config['asterlink']['endpoint_url']) || 
            !$current_user->asterlink_ext_c) {
            return new Response(null, 204);
        }

        $base64url_encode = function($data) {
            return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
        };

        $header = $base64url_encode(json_encode([ 'typ' => 'JWT','alg' => 'HS256' ]));
        $payload = $base64url_encode(json_encode([ 'id' => $security->getUser()->getId()]));
        $signature = $base64url_encode(hash_hmac('sha256', $header.'.'.$payload, $sugar_config['asterlink']['endpoint_token'], true));
        $result = [ 
            'token' => $header.'.'.$payload.'.'.$signature, 
            'endpoint_url' => $sugar_config['asterlink']['endpoint_url']
        ];

        $result['relations'] = [];

        // whate have I done ...
        foreach ($sugar_config['asterlink']['relationships'] as $rel_config) {
            $result['relations'][$rel_config['module']] = [
                'show_create' => $rel_config['show_create'],
                'phone_field' => $rel_config['phone_fields'][0]
            ];
        }
        
        return $result;
    }
}
