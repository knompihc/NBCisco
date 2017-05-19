$(document).ready(function() {
    $body = $("body");
    $('.marow').height($(window).height());
    //$('#page-wrapper').height($(window).height()+5);
    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });

    var host = window.document.location.hostname;
	var port = location.port;
    $('.close').click(function () {
        $(this).parent().addClass('hidden')
    });
    function scrollToAnchor(aid){
        var aTag = $("div[name='"+ aid +"']");
        $('html,body').animate({scrollTop: aTag.offset().top},'slow');
    }
    $("#shg").click(function() {
        $(this).addClass('hidden');
        $("#shc").removeClass('hidden')
        scrollToAnchor('show_sgu');
    });
    $("#shc").click(function() {
        $(this).addClass('hidden');
        $("#shg").removeClass('hidden')
        scrollToAnchor('show_scu');
    });
    ajaxsgu(0);
    ajaxscu(0);
    ajaxzone(0);
    ajaxgroup(0);
    $("#scn").click(function(){
        $("#scp").removeClass("hidden");
        var pg=parseInt($("#scpa").attr('value'));
        ajaxscu(pg+10);
    });
    $("#scp").click(function(){
        $("#scn").removeClass("hidden");
        var pg=parseInt($("#scpa").attr('value'));
        ajaxscu(pg-10);
    });
    $("#sgn").click(function(){
        $("#sgp").removeClass("hidden");
        var pg=parseInt($("#sgpa").attr('value'));
        ajaxsgu(pg+10);
    });
    $("#sgp").click(function(){
        $("#sgn").removeClass("hidden");
        var pg=parseInt($("#sgpa").attr('value'));
        ajaxsgu(pg-10);
    });
    $("#zn").click(function(){
        $("#zp").removeClass("hidden");
        var pg=parseInt($("#zpa").attr('value'));
        ajaxzone(pg+10);
    });
    $("#zp").click(function(){
        $("#zn").removeClass("hidden");
        var pg=parseInt($("#zpa").attr('value'));
        ajaxzone(pg-10);
    });
    $("#grn").click(function(){
        $("#grp").removeClass("hidden");
        var pg=parseInt($("#grpa").attr('value'));
        ajaxgroup(pg+10);
    });
    $("#grp").click(function(){
        $("#grn").removeClass("hidden");
        var pg=parseInt($("#grpa").attr('value'));
        ajaxgroup(pg-10);
    });

    function ajaxgroup(cpid){
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://"+host+":"+port+"/configure/groupconfiguresc?pid="+cpid,
            success: function (msg) {
                $("#grpa").attr('value',cpid);
                if(cpid==0) {
                    $("#grp").addClass("hidden");
                }
                if(msg.substr(msg.length - 1)=="y") {
                    msg=msg.slice(0,-1);
                }
                else{
                    $("#grn").addClass("hidden");
                }
                $("#grcontent").html(msg);
                $('.viewgr').click(function (e) {
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    var id = $(this).attr("id");
                    id = id.substr(4);
                    $("#vidgr_" + id).css({
                        position: "absolute",
                        left: $(window).width()/10,
                        'z-index': "9999"
                    });
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/groupview",
                        success: function (data) {
                            $("#vidgr_" + id).html(data);
                            $("#vidgr_" + id).show();
                            $(".vigr").removeClass('hidden');
                            /* $('body').mouseup(function (e) {
                             var container = $(".viz");

                             if (!container.is(e.target) // if the target of the click isn't the container...
                             && container.has(e.target).length === 0) // ... nor a descendant of the container
                             {
                             container.hide();
                             }
                             });*/
                            $('body').bind( "mouseup touchend", function(e){
                                var container = $(".vigr");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });

                        }
                    });
                });
                $('.addgr').click(function(e){
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    var id=$(this).attr("id");
                    id=id.substr(4);
                    $("#vidgr_"+id).css( {position:"absolute",  left: $(window).width()/10,'z-index': "9999"});
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id="+id,
                        url: "http://"+host+":"+port+"/configure/groupaddsc",
                        success: function (data) {
                            $("#vidgr_"+id).html(data);
                            $("#vidgr_"+id).show();
                            $(".vigr").removeClass('hidden');
                            $('body').bind( "mouseup touchend",function (e)
                            {
                                var container = $(".vigr");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });

                            $('.savegr').click(function(){
                                var id=$(this).attr("id");
                                id=id.substr(4);
                                var values = $("input[name=addgr_"+id+"]:checked").map(function()
                                {
                                    return $(this).val();
                                }).get();
                                if(values.length==0){
                                    swal("Empty!", "Please select atleast one schedule to add!", "warning");
                                    return;
                                }
                                var tarr = [];
                                $.each( values, function( key, value ) {
                                    //alert( key + ": " + value );
                                    tarr.push($("select[name=prigr_"+value+"] option:selected").text());
                                });
                                console.log(tarr);
                                $.ajax({	//create an ajax request to load_page.php
                                    context: this,
                                    type: "GET",
                                    data: "ids="+values+"&sid="+id+"&pri="+tarr,
                                    url: "http://"+host+":"+port+"/configure/groupsavesc",
                                    success: function (data) {
                                        $('#suc').addClass( 'hidden' );
                                        $('#dan').addClass( 'hidden' );
                                        if(data=="Saved Successfully!!")
                                        {
                                            swal("Done!", data, "success");
                                            /*$('#suc').removeClass( 'hidden' );
                                             $('#sucmsg').html(data);*/
                                        }
                                        else
                                        {
                                            swal("Error!", data, "error");
                                            /*$('#dan').removeClass('hidden');
                                             $('#failmsg').html(data);*/
                                        }
                                        $(this).parent().parent().parent().parent().hide();
                                    }
                                });
                            });
                        }
                    });
                });
            }
        });
    }
    function ajaxzone(cpid){
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://"+host+":"+port+"/configure/zoneconfiguresc?pid="+cpid,
            success: function (msg) {
                $("#zpa").attr('value',cpid);
                if(cpid==0) {
                    $("#zp").addClass("hidden");
                }
                if(msg.substr(msg.length - 1)=="y") {
                    msg=msg.slice(0,-1);
                }
                else{
                    $("#zn").addClass("hidden");
                }
                $("#zonecontent").html(msg);
                $('.viewz').click(function (e) {
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    var id = $(this).attr("id");
                    id = id.substr(3);
                    $("#vidz_" + id).css({
                        position: "absolute",
                        left: $(window).width()/10,
                        'z-index': "9999"
                    });
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/zoneview",
                        success: function (data) {
                            $("#vidz_" + id).html(data);
                            $("#vidz_" + id).show();
                            $(".viz").removeClass('hidden');
                           /* $('body').mouseup(function (e) {
                                var container = $(".viz");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });*/
                            $('body').bind( "mouseup touchend", function(e){
                                var container = $(".viz");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });

                        }
                    });
                });
                $('.addz').click(function(e){
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    var id=$(this).attr("id");
                    id=id.substr(3);
                    $("#vidz_"+id).css( {position:"absolute",  left: $(window).width()/10,'z-index': "9999"});
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id="+id,
                        url: "http://"+host+":"+port+"/configure/zoneaddsc",
                        success: function (data) {
                            $("#vidz_"+id).html(data);
                            $("#vidz_"+id).show();
                            $(".viz").removeClass('hidden');
                            $('body').bind( "mouseup touchend",function (e)
                            {
                                var container = $(".viz");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });

                            $('.savez').click(function(){
                                var id=$(this).attr("id");
                                id=id.substr(3);
                                var values = $("input[name=addz_"+id+"]:checked").map(function()
                                {
                                    return $(this).val();
                                }).get();
                                if(values.length==0){
                                    swal("Empty!", "Please select atleast one schedule to add!", "warning");
                                    return;
                                }
                                var tarr = [];
                                $.each( values, function( key, value ) {
                                    //alert( key + ": " + value );
                                    tarr.push($("select[name=priz_"+value+"] option:selected").text());
                                });
                                console.log(tarr);
                                $.ajax({	//create an ajax request to load_page.php
                                    context: this,
                                    type: "GET",
                                    data: "ids="+values+"&sid="+id+"&pri="+tarr,
                                    url: "http://"+host+":"+port+"/configure/zonesavesc",
                                    success: function (data) {
                                        $('#suc').addClass( 'hidden' );
                                        $('#dan').addClass( 'hidden' );
                                        if(data=="Saved Successfully!!")
                                        {
                                            swal("Done!", data, "success");
                                            /*$('#suc').removeClass( 'hidden' );
                                             $('#sucmsg').html(data);*/
                                        }
                                        else
                                        {
                                            swal("Error!", data, "error");
                                            /*$('#dan').removeClass('hidden');
                                             $('#failmsg').html(data);*/
                                        }
                                        $(this).parent().parent().parent().parent().hide();
                                    }
                                });
                            });
                        }
                    });
                });
            }
        });
    }
    function ajaxsgu(cpid){
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://"+host+":"+port+"/configure/sguconfigure?pid="+cpid,
        success: function (msg) {
            $("#sgpa").attr('value',cpid);
            if(cpid==0) {
                $("#sgp").addClass("hidden");
            }
            if(msg.substr(msg.length - 1)=="y") {
                msg=msg.slice(0,-1);
            }
            else{
                $("#sgn").addClass("hidden");
            }
            $("#sgucontent").html(msg);
            $('.viewg').click(function (e) {
                e.preventDefault ? e.preventDefault() : e.returnValue = false;
                var id = $(this).attr("id");
                id = id.substr(3);
                $("#vidg_" + id).css({
                    position: "absolute",
                    left: $(window).width()/10,
                    'z-index': "9999"
                });
                $.ajax({	//create an ajax request to load_page.php
                    type: "GET",
                    data: "id=" + id,
                    url: "http://" + host + ":" + port + "/configure/sguview",
                    success: function (data) {
                        $("#vidg_" + id).html(data);
                        $("#vidg_" + id).show();
                        $(".vig").removeClass('hidden');
                        $('body').bind( "mouseup touchend",function (e) {
                            var container = $(".vig");

                            if (!container.is(e.target) // if the target of the click isn't the container...
                                && container.has(e.target).length === 0) // ... nor a descendant of the container
                            {
                                container.hide();
                            }
                        });

                    }
                });
            });
            $('.addg').click(function(e){
                e.preventDefault ? e.preventDefault() : e.returnValue = false;
                var id=$(this).attr("id");
                id=id.substr(3);
                $("#vidg_"+id).css( {position:"absolute",  left: $(window).width()/10,'z-index': "9999"});
                $.ajax({	//create an ajax request to load_page.php
                    type: "GET",
                    data: "id="+id,
                    url: "http://"+host+":"+port+"/configure/sguadd",
                    success: function (data) {
                        $("#vidg_"+id).html(data);
                        $("#vidg_"+id).show();
                        $(".vig").removeClass('hidden');
                        $('body').bind( "mouseup touchend",function (e)
                        {
                            var container = $(".vig");

                            if (!container.is(e.target) // if the target of the click isn't the container...
                                && container.has(e.target).length === 0) // ... nor a descendant of the container
                            {
                                container.hide();
                            }
                        });

                        $('.saveg').click(function(){
                            var id=$(this).attr("id");
                            id=id.substr(3);
                            var values = $("input[name=addg_"+id+"]:checked").map(function()
                            {
                                return $(this).val();
                            }).get();
                            if(values.length==0){
                                swal("Empty!", "Please select atleast one schedule to add!", "warning");
                                return;
                            }
                            var tarr = [];
                            $.each( values, function( key, value ) {
                                //alert( key + ": " + value );
                                tarr.push($("select[name=prig_"+value+"] option:selected").text());
                            });
                            console.log(tarr);
                            $.ajax({	//create an ajax request to load_page.php
                                context: this,
                                type: "GET",
                                data: "ids="+values+"&sid="+id+"&pri="+tarr,
                                url: "http://"+host+":"+port+"/configure/sgusave",
                                success: function (data) {
                                    $('#suc').addClass( 'hidden' );
                                    $('#dan').addClass( 'hidden' );
                                    if(data=="Saved Successfully!!")
                                    {
                                        swal("Done!", data, "success");
                                        /*$('#suc').removeClass( 'hidden' );
                                        $('#sucmsg').html(data);*/
                                    }
                                    else
                                    {
                                        swal("Error!", data, "error");
                                        /*$('#dan').removeClass('hidden');
                                        $('#failmsg').html(data);*/
                                    }
                                    $(this).parent().parent().parent().parent().hide();
                                }
                            });
                        });
                    }
                });
            });
        }
    });
    }
    function ajaxscu(cpid) {
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/scuconfigure?pid="+cpid,
            success: function (msg) {
                $("#scpa").attr('value',cpid);
                if(cpid==0) {
                    $("#scp").addClass("hidden");
                }
                if(msg.substr(msg.length - 1)=="y") {
                   // $("#scpn").attr('value',npid);
                    msg=msg.slice(0,-1);
                }
                else{
                    $("#scn").addClass("hidden");
                }
                $("#scucontent").html(msg);
                $('.viewc').click(function (e) {
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    var id = $(this).attr("id");
                    id = id.substr(3);
                    $("#vidc_" + id).css({
                        position: "absolute",
                        left: $(window).width()/10,
                        'z-index': "9999"
                    });
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/scuview",
                        success: function (data) {
                            $("#vidc_" + id).html(data);
                            $("#vidc_" + id).show();
                            $(".vic").removeClass('hidden');
                            $('body').bind( "mouseup touchend",function (e) {
                                var container = $(".vic");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });

                        }
                    });
                });
                $('.addc').click(function (e) {
                    var id = $(this).attr("id");
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    id = id.substr(3);
                    $("#vidc_" + id).css({position: "absolute", top: e.pageY-200, left: $(window).width()/10});
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/scuadd",
                        success: function (data) {
                            $("#vidc_" + id).html(data);
                            $("#vidc_" + id).show();
                            $(".vic").removeClass('hidden');
                            $('body').bind( "mouseup touchend",function (e) {
                                var container = $(".vic");

                                if (!container.is(e.target) // if the target of the click isn't the container...
                                    && container.has(e.target).length === 0) // ... nor a descendant of the container
                                {
                                    container.hide();
                                }
                            });

                            $('.savec').click(function () {
                                var id = $(this).attr("id");
                                id = id.substr(3);
                                var values = $("input[name=addc_" + id + "]:checked").map(function () {
                                    return $(this).val();
                                }).get();
                                if(values.length==0){
                                    swal("Empty!", "Please select atleast one schedule to add!", "warning");
                                    return;
                                }
                                var tarr = [];
                                $.each( values, function( key, value ) {
                                    //alert( key + ": " + value );
                                    tarr.push($("select[name=pric_"+value+"] option:selected").text());
                                });
                                $.ajax({	//create an ajax request to load_page.php
                                    context: this,
                                    type: "GET",
                                    data: "ids=" + values + "&sid=" + id+"&pri="+tarr,
                                    url: "http://" + host + ":" + port + "/configure/scusave",
                                    success: function (data) {
                                        $('#suc').addClass('hidden');
                                        $('#dan').addClass('hidden');
                                        if (data == "Saved Successfully!!") {
                                            swal("Done!", data, "success");
                                            /*$('#suc').removeClass('hidden');
                                            $('#sucmsg').html(data);*/
                                        }
                                        else {
                                            swal("Error!", data, "error");
                                            /*$('#dan').removeClass('hidden');
                                            $('#failmsg').html(data);*/
                                        }
                                        $(this).parent().parent().parent().parent().hide();
                                    }
                                });
                            });
                        }
                    });
                });
            }

        });
    }
});
