import { Directive, HostListener, Input } from '@angular/core';
import { AsterlinkService } from '../services/asterlink.service';

@Directive({
    selector: '[stopTelLink]'
})
export class StopTelLinkDirective {
    constructor(private asterlinkService: AsterlinkService) { }

    @HostListener('click', ['$event']) public onClick(event: Event): void {
        if (this.asterlinkService.ready)
            event.preventDefault();
    }
}