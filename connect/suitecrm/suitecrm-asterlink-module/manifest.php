<?php

$manifest = [
    'name' => 'AsterLink',
    'description' => 'Asterisk PBX integration with SuiteCRM',
    'version' => '0.5.1',
    'author' => 'serfreeman1337',
    'readme' => '',
    'acceptable_sugar_versions' => [
        'regex_matches' => ['6\.5\.\d*$'],
    ],
    'icon' => '',
    'is_uninstallable' => true,
    'published_date' => '2024-02-04',
    'type' => 'module',
    'remove_tables' => 'prompt',
];

$installdefs = [
    'id' => 'sf_asterlink',
    'copy' => [
        [
            'from' => '<basepath>/modules/AsterLink',
            'to' => 'modules/AsterLink',
        ],
        [
            'from' => '<basepath>/custom/Extension/application/Ext/EntryPointRegistry/AsterLinkEntryPoint.php',
            'to' => 'custom/Extension/application/Ext/EntryPointRegistry/AsterLinkEntryPoint.php'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Calls/Ext/Vardefs/AsterLink_Override_Calls_duration_hours.php',
            'to' => 'custom/Extension/modules/Calls/Ext/Vardefs/AsterLink_Override_Calls_duration_hours.php'
        ],
        [
            'from' => '<basepath>/custom/Extension/application/Ext/Include/sf_asterlink.php',
            'to' => 'custom/Extension/application/Ext/Include/sf_asterlink.php',
        ]
    ],
    'language' => [
        [
            'from' => '<basepath>/custom/Extension/application/Ext/Language/en_us.AsterLink.php',
            'to_module' => 'application',
            'language' => 'en_us'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Administration/Ext/Language/en_us.AsterLink.php',
            'to_module' => 'Administration',
            'language' => 'en_us'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Calls/Ext/Language/en_us.AsterLink.php',
            'to_module' => 'Calls',
            'language' => 'en_us'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Users/Ext/Language/en_us.AsterLink.php',
            'to_module' => 'Users',
            'language' => 'en_us'
        ],
        [
            'from' => '<basepath>/custom/Extension/application/Ext/Language/ru_RU.AsterLink.php',
            'to_module' => 'application',
            'language' => 'ru_RU'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Administration/Ext/Language/ru_RU.AsterLink.php',
            'to_module' => 'Administration',
            'language' => 'ru_RU'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Calls/Ext/Language/ru_RU.AsterLink.php',
            'to_module' => 'Calls',
            'language' => 'ru_RU'
        ],
        [
            'from' => '<basepath>/custom/Extension/modules/Users/Ext/Language/ru_RU.AsterLink.php',
            'to_module' => 'Users',
            'language' => 'ru_RU'
        ]
    ],
    'custom_fields' => [
        [
            'name' => 'asterlink_uid_c',
            'label' => 'LBL_ASTERLINK_UID',
            'type' => 'varchar',
            'max_size' =>  64,
            'module' => 'Calls',
        ],
        [
            'name' => 'asterlink_cid_c',
            'label' => 'LBL_ASTERLINK_CID',
            'type' => 'phone',
            'max_size' =>  32,
            'module' => 'Calls',
        ],
        [
            'name' => 'asterlink_did_c',
            'label' => 'LBL_ASTERLINK_DID',
            'type' => 'varchar',
            'max_size' =>  32,
            'module' => 'Calls',
        ],
        [
            'name' => 'asterlink_call_seconds_c',
            'label' => 'LBL_ASTERLINK_CALL_SECONDS',
            'type' => 'int',
            'max_size' =>  2,
            'default_value' => 0,
            'module' => 'Calls',
        ],
        [
            'name' => 'asterlink_ext_c',
            'label' => 'LBL_ASTERLINK_EXT',
            'type' => 'varchar',
            'max_size' =>  64,
            'module' => 'Users',
        ],
    ],
    'logic_hooks' => [
        [
            'hook' => 'after_ui_frame',
            'description' => 'asterlink javascript',
            'file' => 'modules/AsterLink/AsterLinkHook.php',
            'class' => 'AsterLink',
            'function' => 'init_javascript'
        ],
    ],
    'administration' => [
        ['from' => '<basepath>/custom/Extension/modules/Administration/Ext/Administration/asterlink_admin.php']
    ],
    'post_execute' => [
        '<basepath>/post_execute.php'
    ],
    'post_uninstall' => [
        '<basepath>/post_uninstall.php'
    ]
];

$upgrade_manifest = [
    'upgrade_paths' => [
        '0.4.0' => [
            'id' => 'sf_asterlink',
            'copy' => [
                [ 'from' => '<basepath>/custom/Extension/application/Ext/EntryPointRegistry/AsterLinkEntryPoint.php', 'to' => 'custom/Extension/application/Ext/EntryPointRegistry/AsterLinkEntryPoint.php' ],
                [ 'from' => '<basepath>/custom/Extension/application/Ext/Include/sf_asterlink.php', 'to' => 'custom/Extension/application/Ext/Include/sf_asterlink.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/AsterLinkHook.php', 'to' => 'modules/AsterLink/AsterLinkHook.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/AsterLinkEntryPoint.php', 'to' => 'modules/AsterLink/AsterLinkEntryPoint.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/AsterLinkEndPoint.php', 'to' => 'modules/AsterLink/AsterLinkEndPoint.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/controller.php', 'to' => 'modules/AsterLink/controller.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/utils.php', 'to' => 'modules/AsterLink/utils.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/language/en_us.lang.php', 'to' => 'modules/AsterLink/language/en_us.lang.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/language/ru_RU.lang.php', 'to' => 'modules/AsterLink/language/ru_RU.lang.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/views/view.config.php', 'to' => 'modules/AsterLink/views/view.config.php' ],
                [ 'from' => '<basepath>/modules/AsterLink/tpls/config.tpl', 'to' => 'modules/AsterLink/tpls/config.tpl' ],
                [ 'from' => '<basepath>/modules/AsterLink/javascript/asterlink.js', 'to' => 'modules/AsterLink/javascript/asterlink.js' ],
                [ 'from' => '<basepath>/modules/AsterLink/javascript/asterlink.worker.js', 'to' => 'modules/AsterLink/javascript/asterlink.worker.js' ]
            ],
            'language' => [
                [ 'from' => '<basepath>/custom/Extension/application/Ext/Language/en_us.AsterLink.php', 'to_module' => 'application', 'language' => 'en_us' ],
                [ 'from' => '<basepath>/custom/Extension/application/Ext/Language/ru_RU.AsterLink.php', 'to_module' => 'application', 'language' => 'ru_RU' ]
            ],
            'post_execute' => [
                '<basepath>/post_execute_040_to_050.php'
            ],
        ]
    ]
];
 
// SuiteCRM 8 support
global $sugar_config;

if (!empty($sugar_config) && strpos($sugar_config['suitecrm_version'], '8.') === 0) {
    unset($installdefs['logic_hooks']);
    $installdefs['copy'][] = [
        'from' => '<basepath>/extensions/asterlink',
        'to' => '../../extensions/asterlink',
    ];

    foreach ($upgrade_manifest['upgrade_paths'] as $ver => &$data) {
        $data['copy'][] = [
            'from' => '<basepath>/extensions/asterlink',
            'to' => '../../extensions/asterlink',
        ];
    }
}

// SuiteCRM module upgrade with zero major version hack. I mean... literally.
if (!empty($_REQUEST['action']) && $_REQUEST['action'] == 'UpgradeWizard_prepare') {
    $uh = new UpgradeHistory();
    $result = $uh->db->query('SELECT id, version FROM '. $uh->table_name . ' WHERE id_name = \'sf_asterlink\' ORDER BY date_entered DESC');

    if (!empty($result)) {
        $temp_version = 0;
        $id = '';

        while ($row = $uh->db->fetchByAssoc($result)) {
            $row['version'] = substr($row['version'], 2); // Strip leading "0." major version and continue check.

            if (!$uh->is_right_version_greater(explode('.', $row['version']), explode('.', $temp_version))) {
                $temp_version = $row['version'];
                $id = $row['id'];
            }
        }

        if ($temp_version) {
            $temp_version = '0.'.$temp_version; // Bring leading "0." back.
            $manifest['description'] = "lol kek cheburek, what is this\"/>
<script>
    $(document).ready(() => {
        $('[name=previous_version]').val('$temp_version');
        $('[name=previous_id]').val('$id');
        $('[name=description]').val('${manifest['description']}');
        $('#why').remove();
    });
</script><input id=why type=hidden ";
        }


    }
}