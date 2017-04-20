function listActivists(activists) {
  $("#activist-list-body").html('');
  var d = document.getElementById('activist-list-body');
  for (var i = 0; i < activists.length; i++) {
    var activist = activists[i];
    var newRow = '<tr>' +
        '<td>' + (i + 1) + '.</td>' +
        '<td>' + activist.name + '</td>' +
        '<td>' + activist.points + '</td>' +
        '<td>' + activist.total_events_30_days + '</td>' +
        '<td>' + activist.total_events + '</td>' +
        '</tr>';
    d.insertAdjacentHTML('beforeend', newRow);
  }
}

function initializeApp() {
  $.ajax({
    url: "/leaderboard/list",
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
