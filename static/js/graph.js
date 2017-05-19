
$(document).ready(function() {
    var host = window.document.location.hostname;
    var port = location.port;
    $body = $("body");

    $(document).on({
        ajaxStart: function() {$body.addClass("loading");    },
        ajaxStop: function() { $body.removeClass("loading"); }
    });

    ajaxcall(0);
    $("#ne").click(function(){
        $('#show').addClass('hidden');
        $("#pr").removeClass("hidden");
        var pg=parseInt($("#pi").attr('value'));
        ajaxcall(pg+10);
    });
    $("#pr").click(function(){
        $('#show').addClass('hidden');
        $("#ne").removeClass("hidden");
        var pg=parseInt($("#pi").attr('value'));
        ajaxcall(pg-10);
    });
    function ajaxcall(cpid) {
        $.ajax({	//create an ajax request to load_page.php
            type: "GET",
            url: "http://" + host + ":" + port + "/configure/graph?pid=" + cpid,
            success: function (msg) {
                $("#pi").attr('value',cpid);
                if(cpid==0) {
                    $("#pr").addClass("hidden");
                }
                if(msg.substr(msg.length - 1)=="y") {
                    msg=msg.slice(0,-1);
                }
                else{
                    $("#ne").addClass("hidden");
                }
                $("#content").html(msg);
                $('input:radio[name=plot]').change(function() {
                    $('#show').removeClass('hidden');
                });
                $("#csv").click(function(){
                    var id=$('input[name=plot]:checked').val();
                    var sel=$("#sel option:selected").text();
                    var sd=$('#datetimepicker').val();
                    var ed=$('#datetimepicker2').val();
                    console.log(sd);
                    console.log(ed);
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        url: "http://" + host + ":" + port + "/configure/csv?id=" + id + "&sel=" + sel + "&sd=" + sd + "&ed=" + ed,
                        success: function (msg) {
                            console.log(msg);
                            var a         = document.createElement('a');
                            a.href        = 'data:attachment/csv,' +  encodeURIComponent(msg);
                            a.target      = '_blank';
                            a.download    = 'reports.csv';
                            document.body.appendChild(a);
                            a.click();
                            /*//window.open("data:application/zip;base64," + content);
                             //var content = "data:text/plain;charset=x-user-defined," + data;
                             var content = "data:application/csv;charset=utf-8; name='reports.csv'," +msg;
                             //var content = "data:application/octet-stream;charset=utf-8" + data;
                             //var content = "data:application/x-zip-compressed;base64," + data;
                             //var content = "data:application/x-zip;charset=utf-8," + data;
                             // var content = "data:application/x-zip-compressed;base64," + data;
                             window.open(content);*/
                            /*var win = window.open("data:application/csv;charset=utf-8,http://" + host + ":" + port + "/"+msg, '_blank');
                             console.log("http://" + host + ":" + port + "/"+msg, '_blank')
                             if(win){
                             //Browser has allowed it to be opened
                             win.focus();
                             }else{
                             //Broswer has blocked it
                             swal('Blocked!','Please allow popups for this site','warning');
                             }*/
                        }
                    });
                });
                $("form").on('submit', function (e) {
                    e.preventDefault();
                    var id=$('input[name=plot]:checked').val();
                    var sel=$("#sel option:selected").text();
                    var sd=$('#datetimepicker').val();
                    var ed=$('#datetimepicker2').val();
                    console.log(sd);
                    console.log(ed);
                    $.ajax({	//create an ajax request to load_page.php
                        type: "GET",
                        url: "http://" + host + ":" + port + "/configure/plot?id=" +id+"&sel="+sel+"&sd="+sd+"&ed="+ed,
                        success: function (msg) {
                            var arr = JSON.parse(msg);
                            var jsonData = [];
                            for (var i = 0; i < arr.length; i++) {
                                var da = {};
                                var arr1 = arr[i].ti.split(/[- :]/),
                                    date = new Date(arr1[0], arr1[1]-1, arr1[2], arr1[3], arr1[4], arr1[5]);
                                da['x']=date;
                                da['y']=(parseFloat(arr[i].val));
                                jsonData.push(da);
                            }
                            console.log(jsonData);
                            /*$('#chart_div').empty();
                             $(window).off('resize');
                             window.morrisObj = NaN;
                             window.morrisObj = new Morris.Line({
                             element: 'chart_div',
                             data: jQuery.parseJSON(msg),
                             xkey: 'ti',
                             ykeys: ['val'],
                             labels: [sel],
                             hideHover: 'auto'
                             });*/
                            $('#graph').removeClass('hidden');
                            Highcharts.setOptions({
                                global: {
                                    useUTC: false
                                }
                            });
                            var chart = new Highcharts.Chart({

                                chart: {
                                    type: 'line',
                                    zoomType: 'xy',
                                    renderTo: 'chart_div',
                                    events: {
                                        load: function() {

                                            // set up the updating of the chart each second
                                            var series = this.series[0];

                                        }}
                                },
                                credits: {
                                    enabled: false
                                },
                                title: {
                                    text: sel+' variation by time'
                                },
                                subtitle: {
                                    text: document.ontouchstart === undefined ?
                                        'Click and drag in the plot area to zoom in' :
                                        'Drag your finger over the plot to zoom in'
                                },
                                xAxis: {
                                    type: 'datetime',
                                    tickPixelInterval: 150,
                                    maxZoom: 5
                                },
                                yAxis: {
                                    labels: {
                                        style: {
                                            color: '#89A54E'
                                        }
                                    },
                                    title: {
                                        text: sel,
                                        style: {
                                            color: '#89A54E'
                                        }
                                    }
                                },

                                tooltip: {
                                    formatter: function() {
                                        return '<b>Time=</b>'+
                                            Highcharts.dateFormat('%Y-%m-%d %H:%M:%S', this.x) +'<br/><b>'+ this.series.name +'=</b>'+
                                            Highcharts.numberFormat(this.y, 2);
                                    },
                                    crosshairs: true
                                },

                                plotOptions: {
                                    line: {
                                        marker: {
                                            enabled: true
                                        }
                                    },
                                    series: {
                                        cursor: 'pointer',
                                        point: {
                                            events: {
                                                click: function (e) {
                                                    hs.htmlExpand(null, {
                                                        pageOrigin: {
                                                            x: e.pageX || e.clientX,
                                                            y: e.pageY || e.clientY
                                                        },
                                                        maincontentText: '<b>Time=</b>'+
                                                        Highcharts.dateFormat('%Y-%m-%d %H:%M:%S', this.x) +'<br/><b>'+ this.series.name +'=</b>'+
                                                        Highcharts.numberFormat(this.y, 2),
                                                        width: 210
                                                    });
                                                }
                                            }
                                        }
                                    }
                                },

                                series: [{
                                    name: sel,
                                    data: jsonData,
                                    color: '#428bca'
                                }]

                            });
                        }
                    });


                    //ajax call here


                    //stop form submission

                });
            }
        });
    }

});
