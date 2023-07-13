<?php

unlink('modules/AsterLink/javascript/c2d.js');
unlink('modules/AsterLink/javascript/ws.js');

// Don't allow SuiteCRM to restore previous version on uninstalling upgraded one.
$backup_path = clean_path(remove_file_extension(urldecode($_REQUEST['install_file'])).'-restore');
rmdir_recursive($backup_path);

$clean_config = loadCleanConfig();

if (!empty($clean_config['asterlink'])) {
    unset($clean_config['asterlink']['endpoint_ws']);
    rebuildConfigFile($clean_config, $clean_config['sugar_version']);
}

if (strpos($clean_config['suitecrm_version'], '8.') === 0) { // Remove logic hook.
    require_once 'include/utils.php';
    require_once 'include/utils/logic_utils.php';

    $hook_array = get_hook_array('');
    if (!empty($hook_array['after_ui_frame'])) {
        foreach ($hook_array['after_ui_frame'] as $hook) {
            if ($hook[3] == 'AsterLink') {
                remove_logic_hook('', 'after_ui_frame', $hook);
                break;
            }
        }
    }
}