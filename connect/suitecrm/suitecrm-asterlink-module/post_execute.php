<?php // serfreeman1337 // 15.06.21 //

// add "Asterisk Extension" field for Users module
$hasField = false;

require_once 'modules/ModuleBuilder/parsers/ParserFactory.php';
$parser = ParserFactory::getParser('editview', 'Users');

foreach ($parser->_viewdefs['panels']['LBL_USER_INFORMATION'] as $row) {
    if (in_array('asterlink_ext_c', $row)) {
        $hasField = true;
        break;
    }
}

if (!$hasField) {
	$parser->_fielddefs['asterlink_ext_c'] = [
			'inline_edit' => 1,
            'required' => false,
            'source' => 'custom_fields',
            'name' => 'asterlink_ext_c',
            'vname' => 'LBL_ASTERLINK_EXT',
            'type' => 'varchar',
            'massupdate' => 0,
            'default' => null,
            'no_default' => false,
            'comments' => '',
            'help' => '',
            'importable' => true,
            'duplicate_merge' => 'disabled',
            'duplicate_merge_dom_value' => 0,
            'audited' => false,
            'reportable' => 1,
            'unified_search' => false,
            'merge_filter' => 'disabled',
            'len' => 64,
            'size' => 20,
            'id' => 'Usersasterlink_ext_c',
            'custom_module' => 'Users',
            'label' => 'LBL_ASTERLINK_EXT'
	];

    $parser->_viewdefs['panels']['LBL_USER_INFORMATION'][] = [
        'asterlink_ext_c',
        '(filler)'
    ];

    $parser->handleSave(false);
}
