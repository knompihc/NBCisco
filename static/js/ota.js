
var host = window.document.location.hostname;
var port = location.port;
var connected=0;
var fname=[];
var own=[];
var bname=[];
var savedbranch="";
var savedrepo="";
var cfname=[];
var cown=[];
var cbname=[];
var csavedbranch="";
var csavedrepo="";
$(document).ready(function() {
    $body = $("body");
    var table=$('#sgutable').DataTable( {
        ajax: "/sguota",
        "columns": [
            { "data": "chk" },
            { "data": "sgu" },
            {"data": "curr" },
            /*{"data": "avail" },*/
            {"data": "status" }
        ],
        "bLengthChange": false,
        columnDefs: [ {
            orderable: false,
            className: 'select-checkbox',
            targets:   0
        } ],
        select: {
            style:    'os',
            selector: 'td:first-child'
        },
        order: [[ 2, 'desc' ]],
        fnCreatedRow: function (nRow, aData, iDataIndex) {
            $(nRow).attr('id', "'"+aData.sgu+"'");
        }
    } );

    var ctable=$('#scutable').DataTable( {
        ajax: "/scuota",
        "columns": [
            { "data": "chk" },
            { "data": "scu" },
            {"data": "curr" },
            /*{"data": "avail" },*/
            {"data": "status" }
        ],
        "bLengthChange": false,
        columnDefs: [ {
            orderable: false,
            className: 'select-checkbox',
            targets:   0
        } ],
        select: {
            style:    'os',
            selector: 'td:first-child'
        },
        order: [[ 2, 'desc' ]],
        fnCreatedRow: function (nRow, aData, iDataIndex) {
            $(nRow).attr('id', "'"+aData.scu+"'");
        }
    } );
    /*$('#all').on('click', function(){
        // Check/uncheck all checkboxes in the table

        var rows = table.cells( ).nodes( );
        $(rows).find('input[type="checkbox"]').prop('checked', this.checked);
    });*/
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/gitconnected?dev=SGU",
        success: function (msg) {
           if(msg=="1"){
               swal('Authenticated','Github Account is Connected!!','success');
            $('#authbtn').html("Re-Authenticate <i class='fa fa-github fa-2x'></i>");
               $('#authbtn').removeClass("btn-success");
               $('#authbtn').addClass("btn-primary");
               $('#cauthbtn').html("Re-Authenticate <i class='fa fa-github fa-2x'></i>");
               $('#cauthbtn').removeClass("btn-success");
               $('#cauthbtn').addClass("btn-primary");
               connected=1;
               $.ajax({	//create an ajax request to load_page.php
                   type: "GET",
                   url: "http://" + host + ":" + port + "/getallrepo?dev=SGU",
                   success: function (msg) {
                       var arr = JSON.parse(msg);
                       var options = $("#reponame");
                       if(arr.length>0){
                           curr=arr[0].id;
                           preval=curr;
                       }
                       $.each(arr, function () {
                           fname[this.fullname] = this.fullname;
                           own[this.fullname] = this.owner;
                           options.append($("<option />").val(this.fullname).text(this.fullname));
                       });
                       $.ajax({	//create an ajax request to load_page.php
                           type: "GET",
                           url: "http://" + host + ":" + port + "/getconnectedrepo?dev=SGU",
                           success: function (msg) {
                               if(msg!="1"&&msg!=null){
                                   var arr = JSON.parse(msg);
                                   $('#reponame').val(arr.git.fullname);
                                   savedbranch=arr.git.branch;
                                   savedrepo=arr.git.fullname;
                                   $('#major').val(arr.major);
                                   $('#minor').val(arr.minor);
                                   $('#name').val(arr.name);
                               }
                               $('#reponame').trigger('change');
                           }
                       });
                   }
               });

               $.ajax({	//create an ajax request to load_page.php
                   type: "GET",
                   url: "http://" + host + ":" + port + "/getallrepo?dev=SCU",
                   success: function (msg) {
                       var arr = JSON.parse(msg);
                       var options = $("#creponame");
                       if(arr.length>0){
                           curr=arr[0].id;
                           preval=curr;
                       }
                       $.each(arr, function () {
                           cfname[this.fullname] = this.fullname;
                           cown[this.fullname] = this.owner;
                           options.append($("<option />").val(this.fullname).text(this.fullname));
                       });
                       $.ajax({	//create an ajax request to load_page.php
                           type: "GET",
                           url: "http://" + host + ":" + port + "/getconnectedrepo?dev=SCU",
                           success: function (msg) {
                               if(msg!="1"&&msg!=null){
                                   var arr = JSON.parse(msg);
                                   $('#creponame').val(arr.git.fullname);
                                   csavedbranch=arr.git.branch;
                                   csavedrepo=arr.git.fullname;
                                   $('#cmajor').val(arr.major);
                                   $('#cminor').val(arr.minor);
                                   $('#cname').val(arr.name);
                               }
                               $('#creponame').trigger('change');
                           }
                       });
                   }
               });
           }else if(msg=="0"){
               swal('ERROR!','Something is not right, not able to authenticate','warning');
           }

        }
    });

    $('#repo').click(function(){
        if(connected!=1){
            swal('Not Connected!','Please Authenticate Github Account!','error');
            return;
        }
        var rurl=$('#reponame').val();
        if(rurl.length==0){
            swal('Warning','Repo cannot be empty!','warning');
            return;
        }
        if($('#branchname').val().length==0||$('#major').val().length==0||$('#minor').val().length==0||$('#name').val().length==0){
            swal('Warning','Everything is required!','warning');
            return;
        }
        $.ajax({
            type: "GET",
            url: "http://" + host + ":" + port + "/connectrepo?dev=SGU&fname="+fname[$('#reponame').val()]+"&owner="+own[$('#reponame').val()]+"&branch="+$('#branchname').val()+"&major="+$('#major').val()+"&minor="+$('#minor').val()+"&name="+$('#name').val(),
            success: function (msg) {
                if(msg=="1"){
                    swal('Connected!','Repository Connected','success');
                }else if(msg=="0"){
                    swal('ERROR!','Something is not right, not able to connect','warning');
                }

            }
        });

    });

    $("#reponame").change(function() {
        $.ajax({
            type: "GET",
            url: "http://" + host + ":" + port + "/getallbranchesforRepo?dev=SGU&fname="+fname[$('#reponame').val()]+"&owner="+own[$('#reponame').val()],
            success: function (msg) {
                var arr = JSON.parse(msg);
                var options = $("#branchname");
                options.html("");
                if(arr.length>0){
                    curr=arr[0].id;
                    preval=curr;
                }
                $.each(arr, function () {
                    bname[this.name] = this.name;
                    options.append($("<option />").val(this.name).text(this.name));
                });
                if(savedbranch!=""&&savedrepo==$('#reponame').val()){
                    $('#branchname').val(savedbranch);
                }
            }
        });
    });

    $("#creponame").change(function() {
        $.ajax({
            type: "GET",
            url: "http://" + host + ":" + port + "/getallbranchesforRepo?dev=SCU&fname="+cfname[$('#creponame').val()]+"&owner="+cown[$('#creponame').val()],
            success: function (msg) {
                var arr = JSON.parse(msg);
                var options = $("#cbranchname");
                options.html("");
                if(arr.length>0){
                    curr=arr[0].id;
                    preval=curr;
                }
                $.each(arr, function () {
                    cbname[this.name] = this.name;
                    options.append($("<option />").val(this.name).text(this.name));
                });
                if(csavedbranch!=""&&csavedrepo==$('#creponame').val()){
                    $('#cbranchname').val(csavedbranch);
                }
            }
        });
    });

    $('#crepo').click(function(){
        if(connected!=1){
            swal('Not Connected!','Please Authenticate Github Account!','error');
            return;
        }
        var rurl=$('#creponame').val();
        if(rurl.length==0){
            swal('Warning','Repo cannot be empty!','warning');
            return;
        }
        if($('#cbranchname').val().length==0||$('#cmajor').val().length==0||$('#cminor').val().length==0||$('#cname').val().length==0){
            swal('Warning','Everything is required!','warning');
            return;
        }
        $.ajax({
            type: "GET",
            url: "http://" + host + ":" + port + "/connectrepo?dev=SCU&fname="+cfname[$('#creponame').val()]+"&owner="+cown[$('#creponame').val()]+"&branch="+$('#cbranchname').val()+"&major="+$('#cmajor').val()+"&minor="+$('#cminor').val()+"&name="+$('#cname').val(),
            success: function (msg) {
                if(msg=="1"){
                    swal('Connected!','Repository Connected','success');
                }else if(msg=="0"){
                    swal('ERROR!','Something is not right, not able to connect','warning');
                }

            }
        });

    });
    $("#update").click(function () {
        if(connected!=1){
            swal('Not Connected!','Please Authenticate Github Account!','error');
            return;
        }
        var ids=[];
        $('.selected').each(function(){
            if($(this).parent().parent().attr('id')=="sgutable"){
                ids.push($(this).attr('id'));
            }
        });
        console.log(ids);
        if(ids.length==0){
            swal("Warning","Please select atleast one row!!","warning");
            return;
        }
        $.ajax({
            type: "GET",
            url: "http://" + host + ":" + port + "/update_firmware?ids=" +ids ,
            success: function (data) {
                if(data=="0")
                {
                    swal({
                            title: "Updating!",
                            text: "Selected SGUs firmware are being updated!!",
                            type: "success"
                        },
                        function(){
                            location.reload();
                        });
                }
                else
                {
                    swal("Error!", "Something went wrong!!", "error");
                }
            }
        });
    });

    $("#cupdate").click(function () {
        if(connected!=1){
            swal('Not Connected!','Please Authenticate Github Account!','error');
            return;
        }
        var ids=[];
        $('.selected').each(function(){
            if($(this).parent().parent().attr('id')=="scutable"){
                ids.push($(this).attr('id'));
            }
        });
        console.log(ids);
        if(ids.length==0){
            swal("Warning","Please select atleast one row!!","warning");
            return;
        }
        $.ajax({
            type: "GET",
            url: "http://" + host + ":" + port + "/update_firmware_scu?ids=" +ids ,
            success: function (data) {
                if(data=="0")
                {
                    swal({
                            title: "Updating!",
                            text: "Selected SCUs firmware are being updated!!",
                            type: "success"
                        },
                        function(){
                            location.reload();
                        });
                }
                else
                {
                    swal("Error!", "Something went wrong!!", "error");
                }
            }
        });
    });

});

