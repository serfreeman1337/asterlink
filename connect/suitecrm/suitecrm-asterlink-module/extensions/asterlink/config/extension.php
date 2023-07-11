<?php

use Symfony\Component\DependencyInjection\Container;

if (!isset($container)) {
    return;
}

/** @var Container $container */
$extensions = $container->getParameter('extensions') ?? [];

$extensions['asterlink'] = [
    'remoteEntry' => './extensions/asterlink/remoteEntry.js',
    'remoteName' => 'asterlink',
    'enabled' => true,
    'extension_name' => 'AsterLink',
    'extension_uri' =>  'https://github.com/serfreeman1337/asterlink',
    'description' => 'Asterisk PBX integration with SuiteCRM',
    'version' =>  '0.5.0',
    'author' =>  'serfreeman1337',
    'author_uri' =>  'https://github.com/serfreeman1337',
    'license' =>  'MIT'
];

$container->setParameter('extensions', $extensions);