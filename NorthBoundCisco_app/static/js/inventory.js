/*Function For Viewing And Updating Inventories  */

$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading"); },
        ajaxStop: function() { $body.removeClass("loading"); }
    });
    function scrollToAnchor(aid){
        var aTag = $("div[name='"+ aid +"']");
        $('html,body').animate({scrollTop: aTag.offset().top},'slow');
    }
    $("#sa").click(function() {
        $(this).addClass('hidden');
        $("#ad").removeClass('hidden')
        scrollToAnchor('update_inven');
    });
    $("#ad").click(function() {
        $(this).addClass('hidden');
        $("#sa").removeClass('hidden')
        scrollToAnchor('inven_main');
    });
    ajaxinven(0);
    $("#ne").click(function(){
        $("#pr").removeClass("hidden");
        var pg=parseInt($("#pi").attr('value'));
        ajaxinven(pg+5);
    });
    $("#pr").click(function(){
        $("#ne").removeClass("hidden");
        var pg=parseInt($("#pi").attr('value'));
        ajaxinven(pg-5);
    });
	    function ajaxinven(cpid) {
        $.ajax({	//create an ajax request to load GO File
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/viewinventories?pid=" + cpid,
            success: function (data) {

                $("#pi").attr('value',cpid);
                if(cpid==0) {
                    $("#pr").addClass("hidden");
                }
                if(data.substr(data.length - 1)=="y") {
                    data=data.slice(0,-1);
                }
                else{
                    $("#ne").addClass("hidden");
                }
                $("#inven").html(data);
                $('tr').click(function(){
                    $('#t'+$(this).attr('id')).focus();

                });
			  $('.saveinven').click(function(e) {
					e.preventDefault ? e.preventDefault() : e.returnValue = false;
                        var id = $(this).attr("id");
						id=id.substr(6);
                        var values = $("#t"+id+ "").text();
                        $.ajax({//create an ajax request to load Go File
                                    context: this,
                                    type: "GET",
                                    data: "&ids=" + values + "&sid=" + id,
                                    url: "http://" + host + ":" + port + "/configure/updateinven",
                                    success: function (data) {
                                        $('#suc').addClass('hidden');
                                        $('#dan').addClass('hidden');
                                        if (data == "DataSaved Successfuly") {
											swal({  
											title: "Done!", 
											text: data,
											type: "success",  
											closeOnConfirm: false,  
											showLoaderOnConfirm: true, });

                                        }
                                        else {
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

					}
			});
    }
});

/*Function for ADD Inventories  */

$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });
    $("#inventory1").on('submit', function (e) {
        e.preventDefault();
        var name=$('#type').val();
        var desc=$('#desc').val();
        //var qtys=$('#qtys').val();
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/AddInventories?name=" +name+"&description="+desc ,
            success: function (data) {
                if (data == "DataSaved Successfully") {
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
						showLoaderOnConfirm: true, }, 
						function(){   
							location.reload();
						});
                }
			}
        });
    });
});