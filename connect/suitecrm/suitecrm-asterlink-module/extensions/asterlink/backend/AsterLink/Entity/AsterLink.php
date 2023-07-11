<?php

namespace App\Extension\asterlink\backend\AsterLink\Entity;

use ApiPlatform\Core\Annotation\ApiResource;
use ApiPlatform\Core\Annotation\ApiProperty;

use App\Extension\asterlink\backend\AsterLink\Controller\GetAsterLink;

/**
 * @ApiResource(
 *     security="is_granted('ROLE_USER')",
 *     shortName="AsterLink",
 *     itemOperations={
 *         "get_asterlink"={
 *             "path"="/asterlink",
 *             "method"="GET",
 *             "controller"=GetAsterLink::class,
 *             "openapi_context"={
 *                 "summary"="Retrieves token and params for AsterLink.",
 *                 "description"="Retrieves token and params for AsterLink.",
 *                 "parameters"={
 *                     {
 *                         "name"="id",
 *                         "in"="path",
 *                         "required"=false,
 *                         "description"="I think I'm doing api-platform wrong...",
 *                         "type"="integer",
 *                     },
 *                 },
 *                 "responses"={
 *                     "200"={
 *                         "description"="Token for AsterLink.",
 *                         "content"={
 *                             "application/json"={
 *                                 "schema"={
 *                                     "type"="object",
 *                                     "properties"={
 *                                         "token"={
 *                                             "type"="string",
 *                                             "nullable"=true,
 *                                             "description"="JWT"
 *                                         },
 *                                         "endpoint_url"={
 *                                             "type"="string",
 *                                             "nullable"=true,
 *                                             "description"="Endpoint Address"
 *                                         },
 *                                         "endpoint_ws"={
 *                                             "type"="string",
 *                                             "nullable"=true,
 *                                             "description"="Endpoint WebSocket"
 *                                         },
 *                                     }
 *                                 }
 *                             }
 *                         }
 *                     },
 *                 }
 *              },
 *              "read"=false,
 *         },
 *     },
 *     collectionOperations={}
 * )
 */
class AsterLink
{
    public function getId() {
        return 1337;
    }
}