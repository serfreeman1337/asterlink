import { Component} from '@angular/core';
import { BaseFieldComponent, DataTypeFormatter, FieldLogicDisplayManager, FieldLogicManager, LanguageStore, MessageService } from 'core';
import { AsterlinkService } from '../../services/asterlink.service';

@Component({
    selector: 'asterlink-phone-field',
    templateUrl: './phone-field.component.html',
    styleUrls: ['./phone-field.component.css']
})
export class PhoneFieldComponent extends BaseFieldComponent  {
    isDialing = false;

    constructor(
        protected typeFormatter: DataTypeFormatter, 
        protected logic: FieldLogicManager,
        protected logicDisplay: FieldLogicDisplayManager,
        private asterlinkService: AsterlinkService,
        private messageService: MessageService,
        private languageStore: LanguageStore
    ) {
        super(typeFormatter, logic, logicDisplay);
    }

    originate(phone: string) {
        if (!this.asterlinkService.ready)
            return;

        this.isDialing = true;
        this.asterlinkService.originate(phone).subscribe({
            next: () => {
                this.messageService.addSuccessMessage(
                    phone + ' ' + 
                    this.languageStore.getAppString('LBL_ASTERLINK_ORIGINATED')
                );
                this.isDialing = false;
            },
            error: e => {
                this.messageService.error(e);
                this.messageService.addDangerMessageByKey('LBL_ASTERLINK_ORIGFAIL');
                this.isDialing = false;
            }
        });
    }
}