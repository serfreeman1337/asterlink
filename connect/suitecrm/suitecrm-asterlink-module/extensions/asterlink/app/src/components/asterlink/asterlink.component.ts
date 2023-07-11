import { Component, OnInit } from '@angular/core';
import { trigger, state, style, transition, animate } from '@angular/animations';
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
export class AsterLinkComponent implements OnInit {
    constructor(public asterlinkService: AsterlinkService) {}

    ngOnInit() {
    }
}