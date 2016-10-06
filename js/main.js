var first = true;
var startNorth = false;
var stops = {};
var currentLine = "";

function getTrainTime(callback) {
    var stop = $('#stop').val();
    var feed = "L";
    if ("L" != currentLine) {
        feed = "123456S"
    }
    $.get('/svc/subway-api/v1/next-trains/'+feed+'/'+stop,
        function(data) {
            var next;
            var following;
            var brooklyn = $('.toggle').data('toggles');
            if (brooklyn.active) {
                if (data.northbound != null) {
                    next = new Date(data.northbound[0]);
                    following = new Date(data.northbound[1]);
                }
            } else {
                if (data.southbound != null) {
                    next = new Date(data.southbound[0]);
                    following = new Date(data.southbound[1]);
                }
            }

            callback(next, following);
        });
}

function timeoutTrain() {
    getTrainTime(updateClock);
    setTimeout(function() {
            timeoutTrain();
    }, 20000);
}

function updateClock(next, following) {
    if (next == undefined) {
console.log("UNDEDFIIIINED");
        if (!first) {
            $('.clocks').each(function(){
                $(this).countdown('stop');
            });
        }
        $('#nextClock').html('N/A');
        $('#followClock').html('N/A');
        return;
    }
    if (!first) {
        $('#nextClock').countdown(next);
        $('#followClock').countdown(following);
    } else {
        first = false;
        $('.clocks').each(function(){
            var $this = $(this), id = $(this).attr('id');  
            var time = following;
            if (id == 'nextClock') {
                time = next;
            }
            $this.countdown(time, function(event) {
                var format = '%M:%S';
                $(this).html(event.strftime(format));
            });
        });
    }
}

// stop|direction|line
function saveLocation() {
    var stop = $('#stop').val();

    var direction = "south";
    var northbound = $('.toggle').data('toggles');
    if (northbound.active) {
        direction = "north";
    }

    var finalLoc = stop+"|"+direction+"|"+currentLine;    
    localStorage.removeItem("savedstop");
    localStorage.setItem("savedstop", finalLoc);
    setLocName(stop);
}

function setLocName(loc) {
    var locName = $('#stop option[value="'+loc+'"]').html();
    var html = '<span class="sbullet mta-bullet mta-'+currentLine.toLowerCase()+'">'+currentLine;
    html += '</span> '+locName.replace(/\&nbsp;/g,'');
    $('#saved').html(html);
    $('#clear').show();
}

function getLocation() {
    return localStorage.getItem("savedstop");
}

function clearLocation() {
    localStorage.removeItem("savedstop");
    $('#saved').html('');
    $('#clear').hide();
}

function changeLine(line) {
    var currClass = "mta-"+currentLine.toLowerCase();
    var newClass = "mta-"+line.toLowerCase();
    $(".main-section .mta-bullet").each(function() {
            var bullet = $(this);
            if (!bullet.hasClass("sbullet")) {
                bullet.removeClass(currClass);
                bullet.addClass(newClass);
                bullet.html(line);
            }
    });
    currentLine = line;
    changeStops(line);
}

function changeStops(line) {
    var lineInfo = stops[line];
    $('.toggle').toggles({
        text:{on:lineInfo["northbound"],off:lineInfo["southbound"]},
        on: startNorth,
        width:250,
        height:50
    });

    var select = $("select");  
    select.find("option").remove(); 
    $(lineInfo["stops"]).each(function() {
        var opt = document.createElement("option");
        opt.value = this[0];
        var name = this[1];
        var sp = 22 - name.length;
        for(var i = 0; i < sp; i++) {
            name = "&nbsp;" + name;
        }
        opt.innerHTML = name;
        select.append(opt);
    });
}

function getStops(callback) {

    var savedLoc = getLocation();   
    var locData = ["L11","south","L"];
    if (savedLoc) {
        var t = savedLoc.split("|");
        if (t.length == 3) {
            locData = t;
        }
    }
    startNorth = locData[1] == "north";

    currentLine = locData[2];
    var work = function() {
        if (savedLoc) {
            console.log("SAVED");
            console.log(locData);
            setLocName(locData[0]);
            $('#stop').val(locData[0]);
        }
    };

    var savedStops = JSON.parse(localStorage.getItem("stops"));
    if (savedStops) {
       stops = savedStops;
       changeLine(currentLine);
       work();
       callback();
       return; 
    }

    $.get("/data/stops.json", function(data) {
        stops = data;
        changeLine(currentLine);
        work();
        localStorage.setItem("stops", JSON.stringify(stops));
        callback();
    });
}

$(function() {
    if (window.navigator.standalone) {
        $("meta[name='apple-mobile-web-app-status-bar-style']").remove();
    }

    getStops(function(){
        // nav line change
        $('.off-canvas-list li').click(function(event){
            event.preventDefault();
            var newLine = $(this).find(".mta-bullet").html();
            changeLine(newLine);
            getTrainTime(updateClock);
        });
        
        $('.toggle').on('toggle', function(){
            getTrainTime(updateClock);
        });
        var stop = $('#stop');
        stop.change(function(event){
            getTrainTime(updateClock);
        });
        $('#save').click(function(event) {
            event.preventDefault();
            saveLocation();
        });
        $('#clear').click(function(event) {
            event.preventDefault();
            clearLocation();
        });
        
        getTrainTime(updateClock);
        setTimeout(function() {
           timeoutTrain(); 
        }, 20000);
    });
});
