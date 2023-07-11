import { Pipe, PipeTransform } from '@angular/core';
import { LanguageStore } from 'core';

@Pipe({
    name: 'moduleName'
})
export class ModuleNamePipe implements PipeTransform {
    constructor(private languageStore: LanguageStore) {}

    transform(value: string): string {
        return this.languageStore.getAppListString('moduleListSingular')[value] ?? 
            (this.languageStore.getAppListString(['parent_type_display'][value] ?? value));
    }
}