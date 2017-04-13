function listEvents(events) {
  if (events.length === 0) {
    flashMessage("No events from server", true);
    return;
  }

  for (var i = 0; i < events.length; i++) {
    var event = events[i];
    var attendeeString = '';
    for (var j = 0; j < event.attendees.length; j++) {
      attendeeString += event.attendees[j];
      if (j !== event.attendees.length - 1) {
        attendeeString +='<br/>';
      }
    }

    // Now, create the link.
    var eventLink = document.createElement('a');
    $(eventLink).text('edit');
    $(eventLink).attr('href', '/update_event/' + event.event_id);

    // output to new row in table to display
    var newRow = '<tr>' +
        '<td><a class="edit-link" href="' + eventLink + '"><span class="glyphicon glyphicon-pencil"></span></a></td>' +
        '<td nowrap>' + event.event_date + '</td>' +
        '<td><b>' + event.event_name + '</b></td>' +
        '<td nowrap>' + event.event_type + '</td>' +
        '<td nowrap>' + event.attendees.length + '</td>' +
        '<td>' + attendeeString + '</td>' +
        '</tr>';
    var d = document.getElementById('event-list-body');
    d.insertAdjacentHTML('beforeend', newRow);
  }
}

function initializeApp() {
  $.ajax({
    url: "/event/list",
    method: "POST",
    success: function(data) {
      var parsed = JSON.parse(data);
      if (parsed.status === "error") {
        flashMessage("Error: " + parsed.message, true);
        return;
      }
      // status === "success"

      // The data must be a list of events b/c it was successful.
      listEvents(parsed);
    },
    error: function() {
      flashMessage("Error connecting to server.", true);
    },
  });
}
