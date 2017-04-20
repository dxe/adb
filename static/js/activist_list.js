function listActivists(activists) {
  $("#activist-list-body").html('');
  var d = document.getElementById('activist-list-body');
  var m = document.getElementById('modals');
  var activistStatus = ""
  var today = new Date();
  var daysAgo60 = new Date();
  daysAgo60.setDate(daysAgo60.getDate() - 30);
  var daysAgo90 = new Date();
  daysAgo90.setDate(daysAgo90.getDate() - 90);

  for (var i = 0; i < activists.length; i++) {
    var activist = activists[i];

    if (activist.firstevent != "None") {
      var firsteventSplit = activist.firstevent.split('-');
      var firsteventDate = new Date(firsteventSplit[0],firsteventSplit[1]-1,firsteventSplit[2]);
    }

    if (activist.lastevent != "None") {
      var lasteventSplit = activist.lastevent.split('-');
      var lasteventDate = new Date(lasteventSplit[0],lasteventSplit[1]-1,lasteventSplit[2]); 
    }


    if (activist.name == "") {
      activistStatus = "";
    }
    else {
      if (activist.firstevent == "None") {
        activistStatus = '<span class="hide">4</span>No attendance';
      }
      else {
        if (lasteventDate < daysAgo60) {
          activistStatus = '<span class="hide">3</span>Former';
        }
        else if (firsteventDate > daysAgo90  && activist.totalevents < 5) {
          activistStatus = '<span class="hide">2</span>New';
        }
        else {
          activistStatus = '<span class="hide">1</span>Current';
        }
      }
    }

    var modalID = 'modal' + activist.id;
    var modal = '<div class="modal fade" id= ' + modalID + ' tabindex="-1" role="dialog">' +
                '<div class="modal-dialog" role="document"><div class="modal-content"><div class="modal-header">' +
                '<h2 class="modal-title">' + activist.name + '</h5></div>' +
                '<div class="modal-body">' +
                  '<label for="email">Email: </label><input class="form-control" type="text" value="'+ activist.email + '" id="email"><br />' +
                  '<label for="chapter">Chapter: </label><input class="form-control" type="text" value="'+ activist.chapter_id + '" id="chapter"><br />' + 
                  '<label for="phone">Phone: </label><input class="form-control" type="text" value="'+ activist.phone + '" id="phone"><br />' +
                  '<label for="location">Location: </label><input class="form-control" type="text" value="'+ activist.location + '" id="location"><br />' +
                  '<label for="facebook">Facebook: </label><input class="form-control" type="text" value="'+ activist.facebook + '" id="facebook"><br />' +
                  '<label for="core">Core/Staff:&nbsp;</label><input class="form-check-input" type="checkbox" id="core"><br />' +
                  '<label for="exclude">Exclude from Leaderboard:&nbsp;</label><input class="form-check-input" type="checkbox" id="exclude"><br />' +
                  '<label for="pledge">Liberation Pledge:&nbsp;</label><input class="form-check-input" type="checkbox" id="pledge"><br />' +
                  '<label for="globalteam">Global Team Member:&nbsp;</label><input class="form-check-input" type="checkbox" id="globalteam">' +
                '<div class="modal-footer"><button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button><button type="button" class="btn btn-primary">Save changes</button>' +
                '</div></div></div></div>'

    var newRow = '<tr>' +
        '<td>' + '<a class="edit-link" href="#" data-toggle="modal" data-target="#' + modalID + '"><button class="btn btn-default glyphicon glyphicon-pencil"></button></a>' + '</td>' +
        '<td>' + activist.name + '</td>' +
        '<td>' + activist.email + '</td>' +
        '<td>' + activist.phone + '</td>' +
        '<td>' + activist.firstevent + '</td>' +
        '<td>' + activist.lastevent + '</td>' +
        '<td>' + activistStatus + '</td>' +
        '</tr>';
    d.insertAdjacentHTML('beforeend', newRow);
    m.insertAdjacentHTML('beforeend', modal);
  }
  $(function(){ $("#activist-list").tablesorter({ sortList: [[6,0], [1,0]] }); });
}

function initializeApp() {
  $.ajax({
    url: "/activist/list",
    success: function(data) {
      var parsed = JSON.parse(data);
      if (parsed.status === "error") {
        flashMessage("Error: " + parsed.message, true);
        return;
      }
      // status === "success"

      listActivists(parsed);
    },
    error: function() {
      flashMessage("Error connecting to server.", true);
    },
  });
}
