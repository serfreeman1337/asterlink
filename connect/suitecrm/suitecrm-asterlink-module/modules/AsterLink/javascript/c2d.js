// serfreeman1337 // 21.03.2020 //

function alInitFields() {
    $('a[href*="tel"]').off('click');

    $('a[href*="tel"]').click(function(e) {
        e.preventDefault();
        
        var a = $(this);
        if (!a.parent().has('.alProgress').length) {
            a.after('\
                <span class="alProgress" style="margin-left: 5px; color: orange;">\
                    <span class="suitepicon suitepicon-module-calls" style="font-size: 12px; color: orange;"></span>\
                    ...\
                </span>');

            $.ajax({
                url: ASTERLINK_URL+'/originate/',
                type: 'post',
                data: {
                    phone: a.text()
                },
                headers: {
                    'X-Asterlink-Token': ASTERLINK_TOKEN,
                },
                success: () => {
                    a.parent().children('.alProgress').remove();
                }
            });
        }
    });
}