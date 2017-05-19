$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });
    $("#Communicationparams").on('submit', function (e) {
        e.preventDefault();
        var baud_rate=$('#baud').val();
		var data_bits=$('#data').val();
		var parity_bits=$('#parity').val();
		var stop_bits=$('#stop').val();
        //var qtys=$('#qtys').val();
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/Communparams?baud_rate="+baud_rate+"&stop_bits="+ stop_bits+"&parity_bits="+parity_bits+"&data_bits="+data_bits,
            success: function (data) {
                if (data == "Packet data send Successfully") {
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
						showLoaderOnConfirm: true, 
						});
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
    $("#Polingparams").on('submit', function (e) {
        e.preventDefault();
        var packet_enable=$('#packet_enable').val();
		var polling_rate=$('#polling_rate').val();
		var response_rate=$('#response_rate').val();
		var time_out=$('#time_out').val();
		var device_id=$('#device_id').val();
        //var qtys=$('#qtys').val();
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/Polingparams?packet_enable="+packet_enable+"&polling_rate="+polling_rate+"&response_rate="+response_rate+"&time_out="+time_out+"&device_id="+device_id,
            success: function (data) {
                if (data == "Packet data send Successfully") {
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
						showLoaderOnConfirm: true, 
						});
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
    $("#AddEnergyParameter").on('submit', function (e) {
        e.preventDefault();
        var DeviceID=$('#devid').val();
		var Length=$('#len').val();
		var Query=$('#query').val();
        //var qtys=$('#qtys').val();
        $.ajax({	//create an ajax request to load Go File
            type: "GET",
            url: "http://" + host + ":" + port + "/AddEnergyParameter?DeviceID="+DeviceID+"&Length="+Length+"&Query="+Query,
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
						showLoaderOnConfirm: true, 
						});
                }
			}
        });
    });
});