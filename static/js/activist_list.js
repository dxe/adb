function listActivists(activists) {
  $("#activist-list-body").html('');
  var d = document.getElementById('activist-list-body');
  var m = document.getElementById('modals');

  for (var i = 0; i < activists.length; i++) {
    var activist = activists[i];

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
        '<td>' + 'Status' + '</td>' +
        '</tr>';
    d.insertAdjacentHTML('beforeend', newRow);
    m.insertAdjacentHTML('beforeend', modal);
  }
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
