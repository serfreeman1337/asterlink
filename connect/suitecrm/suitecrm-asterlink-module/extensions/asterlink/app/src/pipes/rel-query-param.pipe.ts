import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'relQueryParam'
})
export class RelQueryParam implements PipeTransform {
    transform(val: string, phoneField: string): any {
        return {
            [phoneField]: val
        };
    }
}