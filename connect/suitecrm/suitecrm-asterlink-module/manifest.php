<?php // serfreeman1337 // 11.07.2023 //

$manifest = [
    'name' => 'AsterLink',
    'description' => 'Asterisk PBX integration with SuiteCRM',
    'version' => '0.5.0',
    'author' => 'serfreeman1337',
    'readme' => '',
    'acceptable_sugar_versions' => [
        'regex_matches' => ['6\.5\.\d*$'],
    ],
    'icon' => '',
    'is_uninstallable' => true,
    'published_date' => '2023-07-13',
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

$upgrade_manifest = [];
 
// SuiteCRM 8 support
global $sugar_config;

if (!empty($sugar_config) && strpos($sugar_config['suitecrm_version'], '8.') === 0) {
    unset($installdefs['logic_hooks']);
    $installdefs['copy'][] = [
        'from' => '<basepath>/extensions/asterlink',
        'to' => '../../extensions/asterlink',
    ];
}