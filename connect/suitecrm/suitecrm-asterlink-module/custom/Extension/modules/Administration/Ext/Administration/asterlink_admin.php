<?php // serfreeman1337 // 15.06.21 //

$admin_option_defs = [];
$admin_option_defs['Administration']['AsterLink'] = [
    'Asterlink Connector Configuration',
    'LBL_ASTERLINK_CTITLE',
    'LBL_ASTERLINK_CDESC',
    './index.php?module=AsterLink&action=config',
    'system-settings'
];

$admin_group_header[] = array('LBL_ASTERLINK_STITLE', '', false, $admin_option_defs, '');
