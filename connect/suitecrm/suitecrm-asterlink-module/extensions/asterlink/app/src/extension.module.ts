import { NgModule, ApplicationRef, ComponentRef, EmbeddedViewRef, createComponent, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { takeWhile } from 'rxjs/operators';

import { AppStateStore, FieldRegistry, ImageModule } from 'core';

import { AsterlinkService } from './services/asterlink.service';

import { AsterLinkComponent } from './components/asterlink/asterlink.component';
import { CallCardComponent } from './components/call-card/call-card.component';
import { PhoneFieldComponent } from './fields/phone-field/phone-field.component';

import { StopTelLinkDirective } from './directives/stop-tel-link.directive';
import { TelLinkPipe } from './pipes/tel-link.pipe';
import { AppStringsPipe } from './pipes/app-strings.pipe';
import { ModuleNamePipe } from './pipes/module-name.pipe';
import { CallDurationPipe } from './pipes/call-duration.pipe';
import { RecordLinkPipe } from './pipes/record-link.pipe';
import { RelQueryParam } from './pipes/rel-query-param.pipe';

@NgModule({
    declarations: [
        AsterLinkComponent,
        PhoneFieldComponent,
        CallCardComponent,

        StopTelLinkDirective,

        TelLinkPipe,
        AppStringsPipe,
        ModuleNamePipe,
        CallDurationPipe,
        RecordLinkPipe,
        RelQueryParam
    ],
    imports: [
        CommonModule,
        RouterModule,
        ImageModule
    ],
    providers: []
})
export class ExtensionModule {
    private componentRef?: ComponentRef<AsterLinkComponent>;

    private fieldRegistry = inject(FieldRegistry);
    private appRef = inject(ApplicationRef);
    private asterlinkService = inject(AsterlinkService);
    private appStateStore = inject(AppStateStore);
    
    constructor() {
        // Override phone fields.
        this.fieldRegistry.register('default', 'phone', 'list', PhoneFieldComponent);
        this.fieldRegistry.register('default', 'phone', 'detail', PhoneFieldComponent);

        // Wait for app to load all languages.
        this.appStateStore.initialAppLoading$.pipe(
            takeWhile(v => v, true)
        ).subscribe({
            complete: () => {
                this.init();
            }
        });
    }

    init() {
        this.asterlinkService.ready$.subscribe(v => {
            if (v) { // Attach asterlink to the body.
                this.componentRef = createComponent(AsterLinkComponent, { environmentInjector: this.appRef.injector });
                this.appRef.attachView(this.componentRef.hostView);
                const domElem = (this.componentRef.hostView as EmbeddedViewRef<any>).rootNodes[0] as HTMLElement;
                document.getElementsByTagName('app-root')[0].appendChild(domElem);
            } else {
                if (this.componentRef) {
                    this.appRef.detachView(this.componentRef.hostView);
                    this.componentRef.destroy();
                    this.componentRef = undefined;
                }
            }
        });
    }
}