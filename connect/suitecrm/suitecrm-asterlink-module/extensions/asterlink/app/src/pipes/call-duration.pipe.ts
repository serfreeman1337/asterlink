import { Pipe, PipeTransform } from '@angular/core';
import { timer, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { LanguageStore } from 'core';

@Pipe({
    name: 'callDuration'
})
export class CallDurationPipe implements PipeTransform {
    constructor(private languageStore: LanguageStore) {}

    transform(value: string, isAnswered: boolean): Observable<string> {
        const from = Date.parse(value);

        return timer(0, 1000).pipe(
            map(v => {
                const t = Math.floor((Date.now() - from) / 1000);
                const m = Math.floor(t / 60);
                const s = t - m * 60; 

                let duration = '';
                if (!isAnswered) {
                    duration = this.languageStore.getAppString('LBL_ASTERLINK_W');

                    if (m > 0) {
                        duration += ' ' + m + this.languageStore.getAppString('LBL_ASTERLINK_M');
                    }

                    duration += ' ' + s + this.languageStore.getAppString('LBL_ASTERLINK_S');
                } else {
                    duration += m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0');
                }

                return duration;
            })
        );
    }
}