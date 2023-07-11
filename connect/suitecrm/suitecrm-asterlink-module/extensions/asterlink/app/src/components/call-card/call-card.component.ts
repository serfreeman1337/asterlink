import { Component, Input } from '@angular/core';
import { AsterlinkService } from '../../services/asterlink.service';

@Component({
    selector: 'asterlink-call-card',
    templateUrl: './call-card.component.html',
    styleUrls: ['./call-card.component.css'],
})
export class CallCardComponent {
    @Input() isAnswered: boolean;
    @Input() cid: string;
    @Input() did: string;
    @Input() dir: number;
    @Input() id: string;
    @Input() time: string;
    @Input() relations: any;

    constructor(
        public asterlinkService: AsterlinkService
    ) {
    }
}