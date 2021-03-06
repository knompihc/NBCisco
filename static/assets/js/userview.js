
var host = window.document.location.hostname;
var port = location.port;
$(document).ready(function() {
    $body = $("body");
        var table=$('#usertable').DataTable( {
            ajax: "/alluserview",
        "columns": [
            { "data": "chk" },
            { "data": "user_email" },
            {"data": "admin_op" }
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
            $(nRow).attr('id', "'"+aData.user_email+"'");
        }
    } );
});


$('#deleteuser').click(function(){
    var ids=[];
    $('.selected').each(function(){
        ids.push($(this).attr('id'));
    });
    console.log(ids);
    if(ids.length==0){
        swal("Warning","Please select atleast one row!!","warning");
        return;
    }
    swal({
        title: "Are you sure?",
        text: "Selected Entry will be Deleted!",
        type: "warning",
        showCancelButton: true,
        confirmButtonColor: "#DD6B55",
        confirmButtonText: "Yes, do it!",
        cancelButtonText: "No, cancel!",
        closeOnConfirm: false,
        closeOnCancel: false
    }, function(isConfirm){
        if (isConfirm) {
            $.ajax({
                type: "GET",
                url: "http://" + host + ":" + port + "/deleteuserperson?ids=" +ids ,
                success: function (data) {
                    if(data=="1")
                    {
                        swal({
                                title: "Deleted!",
                                text: "Selected Entry has been deleted!!",
                                type: "success",
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
        } else {
            swal("Cancelled", "Operation Cancelled.", "error");
        }
    });

});
