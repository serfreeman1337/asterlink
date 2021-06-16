<?php // serfreeman1337 // 15.06.21 //

$clean_config = loadCleanConfig();
unset($clean_config['asterlink']);
rebuildConfigFile($clean_config, $sugar_version);

// remove "Asterisk Extension" field from Users module
require_once 'modules/ModuleBuilder/parsers/ParserFactory.php';
$parser = ParserFactory::getParser('editview', 'Users');

foreach ($parser->_viewdefs['panels']['LBL_USER_INFORMATION'] as $i => &$row) {
    if (($col = array_search('asterlink_ext_c', $row)) !== false) {
        $row[$col] = '(filler)';
		$fillers = 0;

		foreach ($row as $field)
			if ($field == '(filler)')
				$fillers ++;

		if ($fillers >= 2)
			unset($parser->_viewdefs['panels']['LBL_USER_INFORMATION'][$i]);

		$parser->handleSave(false);

        break;
    }
}