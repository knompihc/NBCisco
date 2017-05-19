$(document).ready(function() {
    $body = $("body");
    $('.marow').height($(window).height());
    var map=[];
    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });

    var host = window.document.location.hostname;
    var port = location.port;
    ajaxzone(0);
    $("#gn").click(function(){
        $("#gp").removeClass("hidden");
        var pg=parseInt($("#gpa").attr('value'));
        ajaxzone(pg+10);
    });
    $("#gp").click(function(){
        $("#gn").removeClass("hidden");
        var pg=parseInt($("#gpa").attr('value'));
        ajaxzone(pg-10);
    });
    $("#add_g").on('submit', function (e) {
        e.preventDefault();
        var name=$('#g_name').val();
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/addgroup?name=" +name ,
            success: function (data) {
                if(data=="done")
                {
                    swal({
                            title: "Done!",
                            text: "Group Added Successfully!!",
                            type: "success",
                            closeOnConfirm: false,
                            showLoaderOnConfirm: true, },
                        function(){
                            location.reload();
                        });
                    /*swal("Done!", "Zone Added Successfully!!", "success",function(){
                     location.reload();
                     }));*/
                }
                else if(data=="already")
                {
                    swal("Duplicate!", "Group Already Present!!", "warning");
                }
                else
                {
                    swal("Error!", data, "error");
                }
            }
        });
    });
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/mapview/getgroup",
        success: function (data) {
            var arr = JSON.parse(data);
            console.log(arr);
            var options = $("#group");
            if(arr.length>0){
                curr=arr[0].id;
                preval=curr;
            }
            $.each(arr, function() {
                map[this.id]=this.name;
                options.append($("<option />").val(this.id).text(this.id));
            });
            if (arr.length>0) {
                $('#group_name').val(map[$('#group').val()]);
            }
        }
    });
    $("#group").change(function() {
        $('#group_name').val(map[$('#group').val()]);
        //initMap($('#zone').val());
    });
    $("#edi_g").on('submit', function (e) {
        e.preventDefault();
        var act=$(document.activeElement).attr('id');
        if (act=="update" && $('#group_name').val()!=map[$('#group').val()])
        {
            $.ajax({
                type: "GET",
                url: "http://" + host + ":" + port + "/configure/updategroup?id=" +$('#group').val()+"&name="+$('#group_name').val() ,
                success: function (data) {
                    if(data=="done")
                    {
                        swal({
                                title: "Updated!",
                                text: "Group has been updated!!",
                                type: "success",
                            },
                            function(){
                                location.reload();
                            });
                        /*swal("Done!", "Zone Added Successfully!!", "success",function(){
                         location.reload();
                         }));*/
                    }
                    else if(data=="already")
                    {
                        swal("Duplicate!", "Group with same name already Present!!", "warning");
                    }
                    else
                    {
                        swal("Error!", data, "error");
                    }
                }
            });
        }
        else if(act=="remove")
        {
            swal({
                title: "Are you sure?",
                text: "Group will be deleted permanently!",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "Yes, delete it!",
                cancelButtonText: "No, cancel!",
                closeOnConfirm: false,
                closeOnCancel: false
            }, function(isConfirm){
                if (isConfirm) {
                    $.ajax({
                        type: "GET",
                        url: "http://" + host + ":" + port + "/configure/removegroup?id=" +$('#group').val() ,
                        success: function (data) {
                            if(data=="done")
                            {
                                swal({
                                        title: "Deleted!",
                                        text: "Group has been deleted!!",
                                        type: "success",
                                    },
                                    function(){
                                        location.reload();
                                    });
                                /*swal("Done!", "Zone Added Successfully!!", "success",function(){
                                 location.reload();
                                 }));*/
                            }
                            else
                            {
                                swal("Error!", data, "error");
                            }
                        }
                    });
                } else {
                    swal("Cancelled", "Operation Cancelled.", "error");
                }
            });
        }
    });

    function ajaxzone(cpid) {
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/groupconfigure?pid="+cpid,
            success: function (msg) {
                $("#gpa").attr('value',cpid);
                if(cpid==0) {
                    $("#gp").addClass("hidden");
                }
                if(msg.substr(msg.length - 1)=="y") {
                    // $("#scpn").attr('value',npid);
                    msg=msg.slice(0,-1);
                }
                else{
                    $("#gn").addClass("hidden");
                }
                $("#groupcontent").html(msg);
                $('.viewc').click(function (e) {
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    var id = $(this).attr("id");
                    id = id.substr(3);
                    $("#vidc_" + id).css({
                        position: "absolute",
                        left: $(window).width()/10,
                        top: e.pageY-180,
                        'z-index': "9999"
                    });
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/groupscuview",
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
                    $("#vidc_" + id).css({position: "absolute", top: e.pageY-180, left: $(window).width()/10,'z-index': "9999"});
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/groupscuadd",
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
                                    swal("Empty!", "Please select atleast one scu to add!", "warning");
                                    return;
                                }
                                $.ajax({	//create an ajax request to load_page.php
                                    context: this,
                                    type: "GET",
                                    data: "ids=" + values + "&gid=" + id,
                                    url: "http://" + host + ":" + port + "/configure/groupscusave",
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
                $('.removec').click(function (e) {
                    var id = $(this).attr("id");
                    e.preventDefault ? e.preventDefault() : e.returnValue = false;
                    id = id.substr(3);
                    $("#vidc_" + id).css({position: "absolute", top: e.pageY-180, left: $(window).width()/10,'z-index': "9999"});
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        data: "id=" + id,
                        url: "http://" + host + ":" + port + "/configure/groupscuremove",
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

                            $('.saverc').click(function () {
                                var id = $(this).attr("id");
                                id = id.substr(3);
                                var values = $("input[name=removec_" + id + "]:checked").map(function () {
                                    return $(this).val();
                                }).get();
                                if(values.length==0){
                                    swal("Empty!", "Please select atleast one scu to remove!", "warning");
                                    return;
                                }
                                $.ajax({	//create an ajax request to load_page.php
                                    context: this,
                                    type: "GET",
                                    data: "ids=" + values + "&gid=" + id,
                                    url: "http://" + host + ":" + port + "/configure/groupscusaver",
                                    success: function (data) {
                                        $('#suc').addClass('hidden');
                                        $('#dan').addClass('hidden');
                                        if (data == "Saved Successfully!!") {
                                            swal("Done!", "Removed Successfully!!", "success");
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
