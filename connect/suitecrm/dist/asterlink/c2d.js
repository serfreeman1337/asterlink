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

            $.post(ASTERLINK_URL+'/originate/', {
                token: ASTERLINK_TOKEN,
                user: ASTERLINK_USER,
                phone: a.text()
            }, function() {
                a.parent().children('.alProgress').remove();
            });
        }
    });
}