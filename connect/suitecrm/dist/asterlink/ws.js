// serfreeman1337 // 09.04.2020 //

function alWs() {
	// TODO: improve security
	let ws = new WebSocket(ASTERLINK_WS + '/ws/?token=' + ASTERLINK_TOKEN + '&user=' + ASTERLINK_USER);

	ws.onmessage = function (e) {
		var d = JSON.parse(e.data);
		d.data.time = new Date(d.data.time);

		d.show ? alShowCard(d.data) : alRemoveCard(d.data);
	};

	ws.onclose = function (e) {
		// reconnect in 1s
		setTimeout(alWs, 1000);
	};

	ws.onerror = function (err) {
		ws.close();
	};
}

const alDirDesc = [["Incoming call", "From"], ["Outgoing call", "To"]];

function alShowCard(d) {
	let c = $('.alCard[data-id="' + d.id + '"]');

	if (!c.length) { // create new card
		c = $('<div />', {
			style: 'display: none; width: 330px;height: 110px;position: fixed;top: 80px;right: 30px;z-index: 9999;background-color: #534d64;border-radius: 5px;color: #f5f5f5;padding: 5px; font-size:16px;',
			class: 'alCard',
			html: '\
<div style="margin-bottom: 5px;">\
    <img src="themes/SuiteP/images/sidebar/modules/Calls.svg"> <span class="alDid">'+ (d.did ? d.did : '-') + '</span>\
        <div style="float: right;">\
        </div>\
	</div>\
    <div style="margin-bottom: 5px;">\
        <span style="color: #f08377;"><a href="#" class="alCall">'+ alDirDesc[d.dir][0] + '</a></span>\
        <span style="color: #e56455; float: right;" class="alDuration"></span>\
</div>\
<table style="width: 100%; font-size:16px;">\
    <tr>\
        <td class="alDirection">'+ alDirDesc[d.dir][1] + '</td>\
        <td class="alCallerId">'+ d.cid + '</td>\
    </tr>\
    <tr>\
        <td>Contact:</td>\
        <td><a class="alContactInfo" href="#"></a></td>\
    </tr>\
</table>', // TODO: make it customisable
			'data-id': d.id
		});

		// enable ajax link for call record
		// TODO: check with disabled ajax
		let callId = c.find('.alCall');

		if (d.id) {
			callId.on('click', () => {
				SUGAR.ajaxUI.go('index.php?module=Calls&action=DetailView&record=' + d.id + '');
			});
		}

		c.appendTo('body');
		c.fadeIn(100);
	}

	// TODO: rewrite / check with disabled ajax
	let contactInfo = c.find('.alContactInfo');
	contactInfo.text(d.contact.id ? d.contact.name : '-- create --');
	contactInfo.off();

	if (d.contact.id) {
	    // view contact
	    contactInfo.on('click', () => {
	        SUGAR.ajaxUI.go('index.php?module=Contacts&action=DetailView&record=' + d.contact.id + '');
	    });
	} else {
	    // create contact
	    contactInfo.on('click', () => {
	        sessionStorage.setItem('alNewContact', d.cid);
	        SUGAR.ajaxUI.go('index.php?module=Contacts&action=EditView');

	        // wait until new contact form is laoded
	        let ltimer = setInterval(() => {
	            if ((SUGAR.ajaxUI.lastCall && !SUGAR.ajaxUI.lastCall.conn) && SUGAR.ajaxUI.lastURL == 'index.php?module=Contacts&action=EditView') {
	                // insert callerid into new contact form
	                $('#phone_work').val(d.cid);
	                sessionStorage.removeItem('alNewContact');
	                clearInterval(ltimer);
	            }
	        }, 100);
	    });
	}

	// TODO: rewrite / check work with time diff
	if (c.data('timer')) {
		clearInterval(c.data('timer'));
	}

	if (d.did) {
		c.find('.alDid').text(d.did);
	}

	// duration updater
	let timer = setInterval(() => {
		let date = +new Date;
		let duration = new Date(date - d.time);
		let durationStr = '';

		if (!d.answered) {
			durationStr = 'awaiting';

			if (duration.getUTCMinutes()) {
				durationStr += ' ' + duration.getUTCMinutes() + 'm.';
			}

			durationStr += ' ' + duration.getSeconds() + 's.';
		} else {
			durationStr += ('00' + duration.getUTCMinutes()).slice(-2) + ':' + ('00' + duration.getSeconds()).slice(-2);
		}

		c.find('.alDuration').text(durationStr);
	}, 100);

	c.data('timer', timer);
}

function alRemoveCard(d) {
	$('.alCard[data-id="' + d.id + '"]').fadeOut(100, function () {
		clearInterval($(this).data('timer'));
		$(this).remove();
	});
}

function alFillContact() {
	// new contact form loaded without ajax
	if (SUGAR.ajaxUI.lastURL == 'index.php?module=Contacts&action=EditView' && sessionStorage.getItem('alNewContact')) {
		// insert callerid into new contact form
		$('#phone_work').val(sessionStorage.getItem('alNewContact'));
		sessionStorage.removeItem('alNewContact');
	}
}