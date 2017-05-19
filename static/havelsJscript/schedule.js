
function validateDate(){
	var fromDate = document.getElementById("datetimepicker").value+" "+document.getElementById("datetimepicker3").value;
	var toDate = document.getElementById("datetimepicker2").value+" "+document.getElementById("datetimepicker4").value;
	if(Date.parse(fromDate) >=Date.parse(toDate)){
		swal("Error!", "Invalid Date Range", "error");
		return false;
	}
	else {
		return true;
	}
}
$(document).ready(function(){
	var host = window.document.location.hostname;
	var port = location.port;
	$("#sch").on('submit', function (e) {
		if (!validateDate()){
			return;
		}
		e.preventDefault();
		var name=$('#sch').serialize();
		$.ajax({	//create an ajax request to load_page.php
			type: "POST",
			url: "http://" + host + ":" + port + "/AddSchedule" ,
			data:name,
			success: function (msg) {
				if(msg=="Schedule Added Successfully!"){
					swal("Done!", msg, "success");
				}
				else if(msg=="Something Went Wrong!")
				{
					swal("Error!", msg, "error");
				}else
				{
					swal("Duplicate!", msg, "warning");
				}
			}
		});
	});
});