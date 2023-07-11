import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'telLink'
})
export class TelLinkPipe implements PipeTransform {
    transform(value: string): string {
        if (!value || value == '')
            return null;
        
        return 'tel:' + value;
    }
}