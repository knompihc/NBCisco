$(document).ready(function() {
    $body = $("body");
    var map1=[];
    var map2=[];
    var map3=[];
    var map4=[];
    var map5=[];
    var map6=[];
    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });

    var host = window.document.location.hostname;
    var port = location.port;
    $.ajax({	//create an ajax request to load_page.php
        type: "GET",
        url: "http://" + host + ":" + port + "/configure/getdeploymentparameter",
        success: function (data) {
            var arr = JSON.parse(data);
            var options = $("#prmid");
            if(arr.length>0){
                curr=arr[0].deployment_id;
                preval=curr;
            }

            $.each(arr, function () {
                map1[this.deployment_id] = this.scu_onoff_pkt_delay;
                map2[this.deployment_id] = this.scu_poll_delay;
                map3[this.deployment_id] = this.scu_schedule_pkt_delay;
                map4[this.deployment_id] = this.scu_onoff_retry_delay;
                map5[this.deployment_id] = this.scu_max_retry;
                map6[this.deployment_id] = this.server_pkt_ack_delay;
                options.append($("<option />").val(this.deployment_id).text(this.deployment_id));
            });

            if (arr.length>0) {
                $('#pktdelay').val(map1[$('#prmid').val()]);
                $('#poldelay').val(map2[$('#prmid').val()]);
                $('#scldpktdelay').val(map3[$('#prmid').val()]);
                $('#rtrydelay').val(map4[$('#prmid').val()]);
                $('#maxrtry').val(map5[$('#prmid').val()]);
                $('#serverpktackdelay').val(map6[$('#prmid').val()]);
            }
        }
    });

    $("#prmid").change(function() {
        $('#pktdelay').val(map1[$('#prmid').val()]);
        $('#poldelay').val(map2[$('#prmid').val()]);
        $('#scldpktdelay').val(map3[$('#prmid').val()]);
        $('#rtrydelay').val(map4[$('#prmid').val()]);
        $('#maxrtry').val(map5[$('#prmid').val()]);
        $('#serverpktackdelay').val(map6[$('#prmid').val()]);
    });
    $("#edi_d").on('submit', function (e) {
        e.preventDefault();
        var prmid=$('#prmid').val();
        var pktdelay = $('#pktdelay').val();
        var poldelay = $('#poldelay').val();
        var scldpktdelay = $('#scldpktdelay').val();
        var rtrydelay = $('#rtrydelay').val();
        var maxrtry = $('#maxrtry').val();
        var serverpktackdelay = $('#serverpktackdelay').val();
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/updatedeploymentparameter?deployment_prm_id=" +prmid+"&deployment_pkt_delay="+pktdelay+"&deployment_pol_delay="+poldelay+"&deployment_sch_delay="+scldpktdelay+"&deployment_rtry_delay="+rtrydelay+"&deployment_max_try="+maxrtry+"&server_pkt_ack_delay="+serverpktackdelay ,
            success: function (msg) {
                if(msg=="done"){
                    swal({
                            title: "Done!",
                            text: "Deployment Parameter has been updated!!",
                            type: "success",
                        },
                        function(){
                            location.reload();
                        });
                    //swal("Done!","Van Added!!","success");
                } else {
                    swal("Error!","Something Went Wrong","Error");
                }
            }
        });
    });
});