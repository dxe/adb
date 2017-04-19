function listActivists(activists) {
  $("#activist-list-body").html('');
  var d = document.getElementById('activist-list-body');
  var m = document.getElementById('modals');

  for (var i = 0; i < activists.length; i++) {
    var activist = activists[i];

    var modalID = 'modal' + activist.id;

    var modal = '<div class="modal fade" id= ' + modalID + ' tabindex="-1" role="dialog">' +
                '<div class="modal-dialog" role="document"><div class="modal-content"><div class="modal-header">' +
                '<h2 class="modal-title">' + activist.name + '</h5>' +
                '<button type="button" class="close" data-dismiss="modal" aria-label="Close"></button></div>' +
                '<div class="modal-body"><b>Email: </b>' + activist.email + '<br /><b>Chapter: </b>' + activist.chapter_id + '<br /><b>Phone: </b>' + activist.phone + '<br /><b>Location: </b>' + activist.location + '<br /><b>Facebook: </b>' + activist.facebook + ' </div>' +
                '<div class="modal-footer"><button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button><button type="button" class="btn btn-primary">Save changes</button>' +
                '</div></div></div></div>'

    var newRow = '<tr>' +
        '<td>' + '<a class="edit-link" href="#" data-toggle="modal" data-target="#' + modalID + '"><button class="btn btn-default glyphicon glyphicon-pencil"></button></a>' + '</td>' +
        '<td>' + activist.name + '</td>' +
        '<td>' + activist.email + '</td>' +
        '<td>' + activist.phone + '</td>' +
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
