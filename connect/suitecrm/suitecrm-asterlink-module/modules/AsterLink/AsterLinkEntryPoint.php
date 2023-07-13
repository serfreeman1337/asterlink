<?php // serfreeman1337 // 13.07.2023 //

if(!defined('sugarEntry') || !sugarEntry) die('Not A Valid Entry Point');

require_once('modules/AsterLink/utils.php');

if (!isset($_SERVER['HTTP_X_ASTERLINK_TOKEN']) || 
    !isset($_REQUEST['action']) ||
    !($config = getConfig())
) {
    http_response_code(400);
    die();
}

if ($_SERVER['HTTP_X_ASTERLINK_TOKEN'] != $config['endpoint_token']) {
    sleep(5);
    http_response_code(403);
    die();
}

$action = $_REQUEST['action'];
$response = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $_SERVER['CONTENT_TYPE'] === 'application/json') {
    $_POST = json_decode(file_get_contents('php://input'), true);
}

switch ($action) {
    case 'get_ext_users':
        $r = BeanFactory::getBean('Users')->get_list("", "asterlink_ext_c IS NOT NULL AND asterlink_ext_c != ''");

        foreach($r['list'] as $user) {
            $response[$user->asterlink_ext_c] = $user->id;
        }
    break;
    case 'create_call_record':
        $response = [];

        $callBean = BeanFactory::newBean('Calls');
        
        foreach ($_POST as $k => $v) {
            $callBean->{$k} = $v;
        }

        $callBean->parent_type = 'Accounts'; // Fix SuiteCRM 8 "My Calls" dashlet explosion.

        $callBean->save();
        $response['id'] = $callBean->id;

        if (!isset($config['relationships']))
            return false;

        foreach ($config['relationships'] as $rel_config) {
            $module_t = strtolower($rel_config['module']);

            $fields = [];

            foreach ($rel_config['phone_fields'] as $phone_field) {
                $fields[] = "`$module_t`.`$phone_field` = '{$callBean->asterlink_cid_c}'";
            }

            $rel = BeanFactory::getBean($rel_config['module'])->get_list("", implode(' OR ', $fields), 0, 1);
            
            if (!$rel['row_count']) {
                continue;
            }

            $rel = $rel['list'][0];

            $callBean->parent_type = $rel_config['module'];
            $callBean->parent_id = $rel->id;
            $callBean->save();

            if (!isset($rel_config['is_parent']) || !$rel_config['is_parent']) {
                $callBean->load_relationship($rel_config['name']);
                $callBean->{$rel_config['name']}->add($rel);
            } else {
                $callBean->parent_type = $rel_config['module'];
                $callBean->parent_id = $rel->id;
                $callBean->save();

                $relatedBean = BeanFactory::getBean($rel_config['module'], $rel->id);
                $relatedBean->load_relationship($rel_config['name']);
                $relatedBean->{$rel_config['name']}->add($callBean);
            }

            $response['relations'][$rel_config['module']] = [
                'id' => $rel->id,
                'name' => $rel->{$rel_config['name_field']},
                'assigned_user_id' => $rel->assigned_user_id
            ];

            if (isset($config['relate_once']) && $config['relate_once'])
                break;
        }

        http_response_code(201);
    break;
    case 'update_call_record':
        $callBean = BeanFactory::getBean('Calls', $_POST['id']);

        foreach ($_POST['data'] as $k => $v) {
            $callBean->{$k} = $v;
        }

        $callBean->date_start = null; // prevent suitecrm from date fields ("save" function in modules/Calls/Call.php)
        $callBean->save();
    break;
    case 'get_relations':
        if (!isset($config['relationships']))
            break;

        $callBean = BeanFactory::getBean('Calls', $_POST['id']);

        foreach ($config['relationships'] as $rel_config) {
            // parent module special
            if (isset($rel_config['is_parent']) && $rel_config['is_parent']) {
                if (!empty($callBean->parent_id) && $callBean->parent_type == $rel_config['module']) {
                    $response[$rel_config['module']] = [
                        'id' => $callBean->parent_id,
                        'name' => $callBean->parent_name,
                    ];
                }

                continue;
            }

            $callBean->load_relationship($rel_config['name']);
            $rels = $callBean->{$rel_config['name']}->getBeans();

            if (empty($rels))
                continue;

            $rel = reset($rels);

            $response[$rel_config['module']] = [
                'id' => $rel->id,
                'name' => $rel->{$rel_config['name_field']},
                'assigned_user_id' => $rel->assigned_user_id
            ];
        }
    break;
    default:
        http_response_code(400);
    break;
}

if (is_null($response)) {
    http_response_code(204);
    die();
}

header('Content-Type: application/json');
echo json_encode($response);
