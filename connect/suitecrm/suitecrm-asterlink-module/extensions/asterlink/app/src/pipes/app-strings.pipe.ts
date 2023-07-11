import { Pipe, PipeTransform } from '@angular/core';
import { LanguageStore } from 'core';

@Pipe({
    name: 'appStrings'
})
export class AppStringsPipe implements PipeTransform {
    constructor(private languageStore: LanguageStore) {}

    transform(value: string): string {
        return this.languageStore.getAppString(value);
    }
}