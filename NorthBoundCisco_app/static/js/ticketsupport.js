$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });
    $("#supporter").on('submit', function (e) {
        e.preventDefault();
        var sub=$('#subject').val();
		var cat=$('#category').val();
		var mail=$('#support_mail').val();
		var num=$('#contactno').val();
        var desc=$('#description').val();
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/supportWTS?support_sub="+sub+"&support_category="+cat+"&support_email="+ mail+"&support_contact="+num+"&support_desc="+desc,
            success: function (data) {
                if (data == "Your Ticket Successfully Placed") {
					swal({  
						title: "Done!", 
						text: data,
						type: "success",  
						closeOnConfirm: false,  
						showLoaderOnConfirm: true, }, 
						function(){   
							location.reload();
						}); 
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

$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });
    $("#supporter1").on('submit', function (e) {
        e.preventDefault();
		var mail1=$('#mail1').val();
		var pass=$('#oldpassword').val();
		var pass1=$('#newpassword').val();
        var pass2=$('#confirmnewpassword').val();
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/supportCP?support_email="+mail1+"&pass="+pass+"&pass1="+pass1+"&pass2="+pass2,
            success: function (data) {
                if (data == "Password Successfully Updated !!!!") {
					swal({  
						title: "Done!", 
						text: data,
						type: "success",  
						closeOnConfirm: false,  
						showLoaderOnConfirm: true, }, 
						function(){   
							location.reload();
						}); 
                }
                else if (data == "") {
					swal({  
						title: "Error!", 
						text: "Invalid emailID or Password ",
						type: "error",  
						closeOnConfirm: false,  
						showLoaderOnConfirm: true, }); 
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