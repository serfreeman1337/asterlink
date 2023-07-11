import { Pipe, PipeTransform } from '@angular/core';
import { ModuleNavigation } from 'core';

@Pipe({
    name: 'recordLink'
})
export class RecordLinkPipe implements PipeTransform {
    constructor(private navigation: ModuleNavigation) {}

    transform(id: string, module: string): string {
        return this.navigation.getRecordRouterLink(module, id);
    }
}