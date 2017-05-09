$(document).ready(function() {
    $body = $("body");
    var map=[];
    var map1=[];
    var map2=[];
    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });

    var host = window.document.location.hostname;
    var port = location.port;
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/configure/getsculoc",
        success: function (data) {
            var arr = JSON.parse(data);
            var options = $("#scu");
            if(arr.length>0){
                curr=arr[0].id;
                preval=curr;
            }
            $.each(arr, function() {
                map[this.id]=this.lat;
                map1[this.id]=this.lng;
                map2[this.id]=this.name;
                options.append($("<option />").val(this.id).text(this.id));
            });
            if (arr.length>0) {
                $('#lat').val(map[$('#scu').val()]);
                $('#lng').val(map1[$('#scu').val()]);
                $('#name').val(map2[$('#scu').val()]);
            }
        }
    });
    $("#scu").change(function() {
        $('#lat').val(map[$('#scu').val()]);
        $('#lng').val(map1[$('#scu').val()]);
        $('#name').val(map2[$('#scu').val()]);
    });
    $("#edi_z").on('submit', function (e) {
        e.preventDefault();
        var act=$(document.activeElement).attr('id');
        if (act=="update" && ($('#name').val()!=map2[$('#scu').val()]||$('#lat').val()!=map[$('#scu').val()]||$('#lng').val()!=map1[$('#scu').val()]))
        {
            $.ajax({
                type: "GET",
                url: "http://" + host + ":" + port + "/configure/updatesculoc?id=" +$('#scu').val()+"&name="+$('#name').val()+"&lat="+$('#lat').val()+"&lng="+$('#lng').val() ,
                success: function (data) {
                    if(data=="done")
                    {
                        swal({
                                title: "Updated!",
                                text: "SCU Location has been updated!!",
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
        }
    });
    $("#add").on('submit', function (e) {
        e.preventDefault();
        var mail1=$('#mail1').val();
        var pass1=$('#newpassword').val();
        var pass2=$('#confirmnewpassword').val();
        var op=$("input[type='radio'][name='admin_op']:checked").val();
        if (pass1!=pass2){
            swal("Error!!","Password didn't match!!","warning");
            return;
        }
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/adduser?userid="+mail1+"&pass="+pass1+"&admin="+op,
            success: function (data) {
                if (data == "done") {
                    swal({
                            title: "Done!",
                            text: "User Added Successfully!!",
                            type: "success",
                            closeOnConfirm: false,
                            showLoaderOnConfirm: true, },
                        function(){
                            location.reload();
                        });
                }else if(data == "already"){
                    swal({
                            title: "Duplicate!",
                            text: "User Already Exists!!",
                            type: "warning",
                            closeOnConfirm: false,
                            showLoaderOnConfirm: true, }
                    );
                }
                else
                {
                    swal({
                        title: "Error!",
                        text: data,
                        type: "error",
                        closeOnConfirm: false,
                        showLoaderOnConfirm: true, });
                }
            }
        });
    });
});
