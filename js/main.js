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
function addLocation() {
    var stop = $('#stop').val();

    var direction = "south";
    var northbound = $('.toggle').data('toggles');
    if (northbound.active) {
        direction = "north";
    }

    var finalLoc = stop+"|"+direction+"|"+currentLine;

    var favs = getFavs();
    favs.push(finalLoc);
    localStorage.removeItem("savedstops");
    localStorage.setItem("savedstops", favs.join(","));
    setLocName(stop);
    createFavLinks();
}

function getLocName(loc) {
    return $('#stop option[value="'+loc+'"]').html();
}

function setLocName(loc) {
    var locName = getLocName(loc);
    var html = '<span class="sbullet mta-bullet mta-'+currentLine.toLowerCase()+'">'+currentLine;
    html += '</span> '+locName.replace(/\&nbsp;/g,'');
    $('#saved').html(html);
    $('#clear').show();
}

function getFavs() {
    var favs = localStorage.getItem("savedstops");
    if (favs) {
        favs = favs.split(",");
    } else {
        favs = [];
    }
    return favs;
}

function getLocation() {
    var old = localStorage.getItem("savedstop");
    var favs = getFavs();
    // deal with transition from old=>new
    if (old != null) {
        favs.push(old);
        localStorage.removeItem("savedstop");
        localStorage.setItem("savedstops", favs.join(","));
    }
    return favs[0];
}

function clearLocation(index) {
    var favs = getFavs();
    favs.splice(index, 1);
    localStorage.removeItem("savedstop");
    localStorage.setItem("savedstops", favs.join(","));
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

function getStopName(line, stop) {
    var lineInfo = stops[line];
    for (var i = 0; i < lineInfo.stops.length; i++) {
        if (lineInfo.stops[i][0] == stop) {
            return lineInfo.stops[i][1].replace(/\&nbsp;/g,''); 
        }
    }
    return "STOP NOT FOUND"; 
}

function getDirectionName(line, dir) {
    var lineInfo = stops[line];

    var name = "";
    if (dir == "south") {
        name = lineInfo.southbound;
    } else {
        name = lineInfo.northbound;
    }
    if (name == "BrooklynBrooklyn Brdgnbsp;Brdg") {
        name = "Brooklyn Brdg";
    }
    return name;
}

function addFavoriteLink(fav, index) {
    var info = fav.split("|");
    var list = document.createElement("li");
    list.className = "fav-item";
    list.setAttribute("data-info", fav);
    var base = document.createElement("a");
    base.href = "#";
    list.appendChild(base);

    // <span class='mta-bullet mta-l fav-train'>L</span>
    var train = document.createElement("span");
    train.className = "mta-bullet mta-"+info[2].toLowerCase()+" fav-train";
    train.innerHTML = info[2];
    base.appendChild(train);

    var trash = document.createElement("span");
    trash.className = "fav-delete";
    trash.setAttribute("data-fav-index", index);
    var trashIcon = document.createElement("img");
    trashIcon.src = "/images/delete_24px.svg";
    trashIcon.alt = "delete favorite";
    trash.appendChild(trashIcon);
    base.appendChild(trash);

    var loc = document.createElement("span");
    loc.className = "fav-loc";
    loc.innerHTML = getDirectionName(info[2], info[1]) +
        '<span style="display:block; font-size:0.2em">BOUND AT</span>'+
        getStopName(info[2], info[0]);
    base.appendChild(loc);

    var clear = document.createElement("div");
    clear.className = "fav-clear";
    base.appendChild(clear);
    $("#fav-list").append(list);
}


function createFavLinks() {
    $(".fav-item").remove();
    var favs = getFavs();
    for (var i = 0; i < favs.length; i++) {
        addFavoriteLink(favs[i], i);
    }
    $('.fav-delete').click(function(event) {
        event.preventDefault();
        var index = $(this).data('fav-index');

        clearLocation(index);
        createFavLinks();
    });
    $('.fav-item').click(function(event) {
        event.preventDefault();
        var info = $(this).data("info").split("|");
        startNorth = info[1] == "north";
        changeLine(info[2]);
        $('#stop').val(info[0]);
        getTrainTime(updateClock);
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

        // populate favorites
        createFavLinks();

        $('.toggle').on('toggle', function(){
            getTrainTime(updateClock);
        });
        var stop = $('#stop');
        stop.change(function(event){
            getTrainTime(updateClock);
        });
        $('#save').click(function(event) {
            event.preventDefault();
            addLocation();
        });

        getTrainTime(updateClock);
        setTimeout(function() {
           timeoutTrain();
        }, 20000);
    });
});
