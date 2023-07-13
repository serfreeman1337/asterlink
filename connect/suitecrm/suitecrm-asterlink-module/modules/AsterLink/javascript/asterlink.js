// serfreeman1337 // 10.07.2023 //

/** Enable Click-2-Call for phone fields. */
function alInitFields() {
  $('a[href*="tel"]')
      .off('click')
      .click(e  => {
        const a = $(e.target);

        if (!a.parent().has('.al-rogress').length) {
          a.after('<span class="al-rogress suitepicon suitepicon-module-calls" '+
              'style="margin-left: 5px; color: orange; font-size: 12px; color: orange;">' +
              '...</span>'
          );

          let headers = {}

          if (typeof ASTERLINK_TOKEN != 'undefined') {
            headers['X-Asterlink-Token'] = ASTERLINK_TOKEN;
          }

          $.ajax({
            url: ASTERLINK_URL+'originate',
            type: 'post',
            data: {
              phone: a.text()
            },
            headers,
            success: () => {
              a.parent().find('.al-rogress').remove();
            }
          });
        }

        return false;
      });
}

/** AsterLink Call Card. */
class AlCallCard {
  /** @private {boolean} Answered state. */
  answered_ = false;

  /** @private {number|undefined} Timer ID. */
  timer_ = undefined;

  /** @private {JQuery|undefined} Card's div component */
  card_ = undefined;

  /** @private {!number} Call time to count duration from. */
  time_ = 0;

  /**
   * Append new call card to the body.
   */
  constructor() {
    this.card_ = $('<div />', {
        class: 'asterlink-call-card bg-primary',
        style: 'display: none; width: 330px; position: fixed; top: 80px; right: 30px; z-index: 9999; border-radius: 5px; padding: 5px; font-size:16px;',
        html: `
<div>
    <img src="themes/SuiteP/images/sidebar/modules/Calls.svg">
    <span class="al-did"></span>
</div>
    
<div class="al-call-info">
    <a href="#" class="al-record"></a>
    <span class="al-duration" style="float: right;"></span>
</div>
    
<div class="al-call-content" style="display: grid; grid-template-columns: minmax(auto, 1fr) 1fr;">
    <span class="al-dir"></span>
    <span class="al-cid" style="word-break: break-word;"></span>
</div>`
    });
    
    this.card_.appendTo('body');
    this.card_.fadeIn(100);
  }

  /**
   * Removes call card from the document body.
   */
  destroy() {
    this.card_.fadeOut(100, () => {
      clearInterval(this.timer_);
      this.card_.remove();
    });
  }

  /**
   * Update call card with a new data.
   * @param {Object} data New data for call card.
   */
  update(data) {
    for (const [key, val] of Object.entries(data)) {
      switch (key) {
        case 'id':
          const url = 'index.php?module=Calls&action=DetailView&record=' + val;
          this.card_.find('.al-record')
              .attr('href', url)
              .off('click')
              .click(() => {
                SUGAR.ajaxUI.go(url);
                return false;
              });
          break;
        case 'dir':
          this.card_.find('.al-record').text(SUGAR.language.get('app_strings', val == 0 ? 'LBL_ASTERLINK_IN' : 'LBL_ASTERLINK_OUT'));
          this.card_.find('.al-dir').text(SUGAR.language.get('app_strings', val == 0 ? 'LBL_ASTERLINK_FROM' : 'LBL_ASTERLINK_TO'));
          break;
        case 'did':
        case 'cid':
          this.card_.find('.al-' + key).text(val);
          break;
        case 'answered':
          this.answered_ = val;
        case 'time':
          this.time_ = Date.parse(val);
          this.updateDuration_();

          if (this.timer_) clearInterval(this.timer_);
          this.timer_ = setInterval(() => this.updateDuration_(), 1000);
          break;
        case 'relations':
          const cc = this.card_.find('.al-call-content');
          cc.find('.al-rel').remove();

          for (const [module, rel] of Object.entries(ASTERLINK_RELMODULES)) {
            const hasRelation = (val && val[module] != null);

            if (!rel.show_create && !hasRelation)
              continue;

            $('<span />', {
              class: 'al-rel',
              text: SUGAR.language.get('app_list_strings', 'moduleListSingular')[module] ??
                        SUGAR.language.get('app_list_strings', 'parent_type_display')[module] ??
                        module
            }).appendTo(cc);

            const url = hasRelation ?
                'index.php?module=' + module + '&action=DetailView&record=' + val[module].id :
                'index.php?module=' + module + '&action=EditView&' + new URLSearchParams({
                  [rel.phone_field]: this.card_.find('.al-cid').text()
                }).toString();
              
            const name = hasRelation ? 
                val[module].name : 
                SUGAR.language.get('app_strings', 'LBL_ASTERLINK_CREATE');

            $('<a />', {
              class: 'al-rel',
              href: url,
              text: name,
              style: 'word-break: break-word;',
              click: () => {
                SUGAR.ajaxUI.go(url);
                return false;
              }
            }).appendTo(cc);
          }

          break;
      }
    }
  }

  /** @private */
  updateDuration_() {
    const t = Math.floor((Date.now() - this.time_) / 1000);
    const m = Math.floor(t / 60);
    const s = t - m * 60;

    let duration = '';

    if (!this.answered_) {
      duration = SUGAR.language.get('app_strings', 'LBL_ASTERLINK_W');

      if (m > 0) {
        duration += ' ' + m + SUGAR.language.get('app_strings', 'LBL_ASTERLINK_M');
      }

      duration += ' ' + s + SUGAR.language.get('app_strings', 'LBL_ASTERLINK_S');
    } else {
      duration += m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0');
    }

    this.card_.find('.al-duration').text(duration);
  }
}

const alCalls = new Map();

const alWorker = new SharedWorker('modules/AsterLink/javascript/asterlink.worker.js');

alWorker.port.onmessage = e => {
  if (!e.data) { // Worker disconnected from stream.
    for (const [id, call] of alCalls) { // Delete all call cards.
      call.destroy();
      alCalls.delete(id);
    }

    return;
  }

  const { show, data } = e.data;
  
  let call = alCalls.get(data.id);

  if (show) {
    if (!call) { // Construct new call card.
      call = new AlCallCard();
      alCalls.set(data.id, call);
    }

    call.update(data);
  } else {
    if (call) {
      call.destroy();
      alCalls.delete(data.id);
    }
  }
};

alWorker.port.postMessage(ASTERLINK_URL + 'stream' + 
    (typeof ASTERLINK_TOKEN != 'undefined' ? 
        '?token=' + ASTERLINK_TOKEN :
        ''
    )
);