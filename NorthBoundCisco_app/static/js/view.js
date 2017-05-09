$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });
    $("#help").hover(function(){
        $("#vic").css( {position:"absolute", top:event.pageY, left: event.pageX,'z-index': "9999",display:'block'});
        $("#vic").removeClass('hidden');
    },function(){
        $("#vic").css( {position:"absolute", top:event.pageY, left: event.pageX,'z-index': "9999",display:'none'});
        $("#vic").addClass('hidden');
    });
    function scrollToAnchor(aid){
        var aTag = $("div[name='"+ aid +"']");
        $('html,body').animate({scrollTop: aTag.offset().top},'slow');
    }
    $("#vi").click(function() {
        $(this).addClass('hidden');
        $("#ad").removeClass('hidden')
        scrollToAnchor('vi');
    });
    $("#ad").click(function() {
        $(this).addClass('hidden');
        $("#vi").removeClass('hidden')
        scrollToAnchor('add');
    });
    ajaxcall(0);
    $("#ne").click(function(){
        $("#pr").removeClass("hidden");
        var pg=parseInt($("#pi").attr('value'));
        ajaxcall(pg+10);
    });
    $("#pr").click(function(){
        $("#ne").removeClass("hidden");
        var pg=parseInt($("#pi").attr('value'));
        ajaxcall(pg-10);
    });
    function ajaxcall(cpid) {
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/viewschedules?pid=" + cpid,
            success: function (msg) {
                $("#pi").attr('value',cpid);
                if(cpid==0) {
                    $("#pr").addClass("hidden");
                }
                if(msg.substr(msg.length - 1)=="y") {
                    msg=msg.slice(0,-1);
                }
                else{
                    $("#ne").addClass("hidden");
                }
                $("#sche").html(msg);
            }
        });
    }
});