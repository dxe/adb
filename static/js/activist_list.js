function listActivists(activists) {
  $("#activist-list-body").html('');
  var d = document.getElementById('activist-list-body');

  for (var i = 0; i < activists.length; i++) {
    var activist = activists[i];

    var newRow = '<tr>' +
        '<td>' + activist.id + '</td>' +
        '<td>' + activist.name + '</td>' +
        '<td>' + activist.email + '</td>' +
        '<td>' + activist.chapter_id + '</td>' +
        '<td>' + activist.phone + '</td>' +
        '<td>' + activist.location + '</td>' +
        '<td>' + activist.facebook + '</td>' +
        '</tr>';
    d.insertAdjacentHTML('beforeend', newRow);
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
