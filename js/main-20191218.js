var first = true;
var startNorth = false;
var stops = {};
var currentLine = "";

function getTrainTime(callback) {
    var stop = $('#stop').val();
    $.get('/svc/subway-api/v1/next-trains/'+currentLine+'/'+stop,
        function(data) {
            var next;
            var following;
            var following2;
            var brooklyn = $('.toggle').data('toggles');
            if (brooklyn.active) {
                if (data.northbound != null) {
                    if (data.northbound.length > 0) {
                        next = new Date(data.northbound[0]);
                    }
                    if (data.northbound.length > 1) {
                        following = new Date(data.northbound[1]);
                    }
                    if (data.northbound.length > 2) {
                        following2 = new Date(data.northbound[2]);
                    }
                }
            } else {
                if (data.southbound != null) {
                    if (data.southbound.length > 0) {
                        next = new Date(data.southbound[0]);
                    }
                    if (data.southbound.length > 1) {
                        following = new Date(data.southbound[1]);
                    }
                    if (data.southbound.length > 2) {
                        following2 = new Date(data.southbound[2]);
                    }
                }
            }
			if (stop == null) {
				return
			}
            var hash = '#'+currentLine+'/'+stop+'/'+brooklyn.active
            if(history.pushState) {
                history.pushState(null, null, hash);
            }
            else {
                location.hash = hash;
            }

            callback(next, following, following2);
        });
}

function timeoutTrain() {
    getTrainTime(updateClock);
    setTimeout(function() {
            timeoutTrain();
    }, 20000);
}

function updateClock(next, following, following2) {
    if (next == undefined) {
        if (!first) {
            $('.clocks').each(function(){
                $(this).countdown('stop');
            });
        }
        $('#nextClock').html('N/A');
        $('#followClock').html('N/A');
        $('#followClock2').html('N/A');
        return;
    }
    if (!first) {
        $('#nextClock').countdown(next);
        if (following) {
            $('#followClock').countdown(following);
        } else {
            $('#followClock').html('N/A');
        }
        if (following2) {
            $('#followClock2').countdown(following2);
        } else {
            $('#followClock2').html('N/A');
        }

    } else {
        first = false;
        $('.clocks').each(function(){
            var $this = $(this), id = $(this).attr('id');
            var time = following;
            if (id == 'nextClock') {
                time = next;
            } else if (id == 'followClock2') {
                time = following2;
            }
            $this.countdown(time, function(event) {
                var format = '%M:%S';
                if (event == undefined) {
                    $(this).html('N/A');
                }
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
				console.log(bullet);
                bullet.removeClass(currClass);
                bullet.addClass(newClass);
				console.log(bullet);
                bullet.html(line);
            }
    });
    currentLine = line;
    changeStops(line);
}

function changeStops(line) {
    var lineInfo = stops[line];
    $('.toggle').toggles({
        text:{on:lineInfo["Northbound"],off:lineInfo["Southbound"]},
        on: startNorth,
        width:250,
        height:50
    });

    var select = $("select");
    select.find("option").remove(); 
    $(lineInfo["Stops"]).each(function() {
        var opt = document.createElement("option");
        opt.value = this.ID;
        var name = this.MTAName;
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

    var savedStops = JSON.parse(localStorage.getItem("stopsv5"));
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
        localStorage.setItem("stopsv5", JSON.stringify(stops));
        callback();
    });
}

function getStopName(line, stop) {
    var lineInfo = stops[line];
    for (var i = 0; i < lineInfo.Stops.length; i++) {
        if (lineInfo.Stops[i].ID == stop) {
            return lineInfo.Stops[i].MTAName.replace(/\&nbsp;/g,''); 
        }
    }
    return "STOP NOT FOUND"; 
}

function getDirectionName(line, dir) {
    var lineInfo = stops[line];

    var name = "";
    if (dir == "south") {
        name = lineInfo.Southbound;
    } else {
        name = lineInfo.Northbound;
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

        if (window.location.hash) {
            var info = window.location.hash.split('/');
            if (info.length > 2) {
                startNorth = info[2] == "true";
                changeLine(info[0].substr(1));
                $('#stop').val(info[1]);
                getTrainTime(updateClock);
            }
        }

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
