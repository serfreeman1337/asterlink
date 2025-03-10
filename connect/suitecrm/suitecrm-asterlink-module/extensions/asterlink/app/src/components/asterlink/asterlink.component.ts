import { Component, inject } from '@angular/core';
import { trigger, style, transition, animate } from '@angular/animations';
import { AsterlinkService } from '../../services/asterlink.service';

@Component({
    selector: 'asterlink',
    templateUrl: './asterlink.component.html',
    animations: [
        trigger('fade', [
            transition('void => *', [
                style({ opacity: 0 }),
                animate(100, style({ opacity: 1 }))
            ]),
            transition('* => void', [
                animate(100, style({ opacity: 0 }))
            ])
        ])
    ],
})
export class AsterLinkComponent {
    public asterlinkService = inject(AsterlinkService);
}