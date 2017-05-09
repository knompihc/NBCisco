$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/auth",
        success: function (msg) {
            if (msg == "0") {
                window.location = "http://" + host + ":" + port + "/login.html";
            }
            else if (msg == "1") {
                window.location = "http://" + host + ":" + port + "/adminlogin.html";
            }
            else if (msg == "2") {
                $('#wrapper').removeClass('hidden');
            }

        }
    });
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/isadmin",
        success: function (msg) {
            if (msg == "1") {
                $("#cfg").removeClass('hidden');
                if($("#cfgi")){
                    $("#cfgi").removeClass('hidden');
                }
            }

        }
    });
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/getuid",
        success: function (msg) {
            if (msg) {
                var bo=$('#uid');
                if(bo){
                    bo.html(msg);
                }
            }

        }
    });
    $('#signout').click(function () {
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/signout",
            success: function (msg) {
                if(msg=="0"){
                    window.location = "http://" + host + ":" + port + "/login.html";
                }
            }
        });
    });
});