// serfreeman1337 // 15.06.21 //

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

const alDirDesc = [
	[SUGAR.language.get('app_strings', 'LBL_ASTERLINK_IN'), SUGAR.language.get('app_strings', 'LBL_ASTERLINK_FROM')],
	[SUGAR.language.get('app_strings', 'LBL_ASTERLINK_OUT'), SUGAR.language.get('app_strings', 'LBL_ASTERLINK_TO')]
];

function alShowCard(d) {
	let c = $('.alCard[data-id="' + d.id + '"]');
	let table;

	if (!c.length) { // create new card
		c = $('<div />', {
			style: 'display: none; width: 330px; position: fixed;top: 80px;right: 30px;z-index: 9999;border-radius: 5px;padding: 5px; font-size:16px;',
			class: 'alCard bg-primary',
			html: '\
<div style="margin-bottom: 5px;">\
    <img src="themes/SuiteP/images/sidebar/modules/Calls.svg"> <span class="alDid">'+ (d.did ? d.did : '-') + '</span>\
        <div style="float: right;">\
        </div>\
	</div>\
    <div style="margin-bottom: 5px;">\
        <a href="#" class="alCall">'+ alDirDesc[d.dir][0] + '</a>\
        <span style="float: right;" class="alDuration"></span>\
</div>', // TODO: make it customisable
			'data-id': d.id
		});

		table = $('<table />', {
			class: 'alDetailsTable',
			style: 'width: 100%; font-size:16px;'
		}).appendTo(c);

		$('<tr />', {
			class: 'alDirection',
			html: '\
<td class="alDirection">' + alDirDesc[d.dir][1] + '</td>\
<td class="alCallerId">'+ d.cid + '</td>'
		}).appendTo(table);

		for (const [module, moduleData] of Object.entries(ASTERLINK_RELMODULES)) {
			if (!moduleData.show_create)
				continue;
			
			$('<tr />', {
				html: '\
<td>'+ moduleData.name + '</td>\
<td><a class="al'+ module + 'Info" href="#" data-module="' + module + '" data-field="' + moduleData.phone_field +'"></a></td>'
			}).appendTo(table);
		}

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
	} else {
		table = $('.alDetailsTable');
	}

	// TODO: rewrite / check with disabled ajax

	if (d.relations) {
		for (let [module, rel] of Object.entries(d.relations)) {
			let url = 'index.php?module='+module+'&action=DetailView&record=' + rel.id;
			let link = table.find('.al'+module+'Info');
	
			if (link.length) {
				link.text(rel.name);
				link.attr('href', url);
				link.off();
	
				
			} else {
				$('<tr />', {
					html: '\
<td>'+ ASTERLINK_RELMODULES[module].name + '</td>\
<td><a class="al'+ module + 'Info" href="'+url+'">'+rel.name+'</a></td>'
				}).appendTo(table);
	
				link = table.find('.al' + module + 'Info');
				link.off();
			}
	
			link.on('click', () => {
				SUGAR.ajaxUI.go(url);
				return false;
			});
		}
	}

	table.find('tr:not(:first-child) a').each(function() {
		let link = $(this);

		if (link.attr('href') != '#') {
			return;
		}

		let url = 'index.php?module='+link.data('module')+'&action=EditView';

		link.text(SUGAR.language.get('app_strings', 'LBL_ASTERLINK_CREATE'));
		link.attr('href', url);
		link.off();

		// create new record
		link.on('click', () => {
			sessionStorage.setItem('alNewRecord', d.cid);
			SUGAR.ajaxUI.go(url);

			// wait until new record form is loaded
			let ltimer = setInterval(() => {
				if ((SUGAR.ajaxUI.lastCall && !SUGAR.ajaxUI.lastCall.conn) && SUGAR.ajaxUI.lastURL == url) {
					// insert callerid into new record form
					$('#'+link.data('field')).val(d.cid);
					sessionStorage.removeItem('alNewRecord');
					clearInterval(ltimer);
				}
			}, 100);

			return false;
		});
	});

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
			durationStr = SUGAR.language.get('app_strings', 'LBL_ASTERLINK_W');

			if (duration.getUTCMinutes()) {
				durationStr += ' ' + duration.getUTCMinutes() + SUGAR.language.get('app_strings', 'LBL_ASTERLINK_M');
			}

			durationStr += ' ' + duration.getSeconds() + SUGAR.language.get('app_strings', 'LBL_ASTERLINK_S');
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
	// TODO: update to work without ajax

	// new contact form loaded without ajax
	// if (SUGAR.ajaxUI.lastURL == 'index.php?module=Contacts&action=EditView' && sessionStorage.getItem('alNewContact')) {
	// 	// insert callerid into new contact form
	// 	$('#phone_work').val(sessionStorage.getItem('alNewContact'));
	// 	sessionStorage.removeItem('alNewContact');
	// }
}