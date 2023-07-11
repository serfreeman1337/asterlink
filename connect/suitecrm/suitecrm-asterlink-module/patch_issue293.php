<?php

// See https://github.com/salesagility/SuiteCRM-Core/issues/293
$content = file_get_contents('../../core/backend/Security/LegacySessionDenyAccessListener.php');

if ($content && strpos($content, 'onSecurityPostValidation') == false) {
    $pos = strrpos($content, '}');
    $pos = strrpos(substr($content, 0, $pos), '}');
    $pos++;

    $content = substr($content, 0, $pos) . '

    /**
     * @param RequestEvent $event
     * @throws ResourceClassNotFoundException
     */
    public function onSecurityPostValidation(RequestEvent $event): void
    {
        $this->decorated->onSecurityPostValidation($event);
        $this->checkLegacySession($event->getRequest(), \'security_post_validation\');
    }' . substr($content, $pos);

    file_put_contents('../../core/backend/Security/LegacySessionDenyAccessListener.php', $content);
}