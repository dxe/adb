function listEvents(events) {
  if (events.length === 0) {
    flashMessage("No events from server");
    return;
  }

  var frag = document.createDocumentFragment();
  for (var i = 0; i < events.length; i++) {
    // This is going to be ugly, but I'm just going to concat
    // everything into one string and shove it into a div. Someone
    // else should make this nicer.
    var event = events[i];
    var text = event.event_name + ' ' + event.event_type + ' ' +
        event.event_date + ': ';
    for (var j = 0; j < event.attendees.length; j++) {
      text += event.attendees[j];
      if (j !== event.attendees.length - 1) {
        text += ', ';
      }
    }
    var div = document.createElement('div');
    $(div).text(text + ' ');

    // Now, create the link.
    var eventLink = document.createElement('a');
    $(eventLink).text('edit');
    $(eventLink).attr('href', '/update_event/' + event.event_id);

    // Now put the link inside the div.
    $(div).append(eventLink);

    // Finally, append the div to the fragment.
    $(frag).append(div);
  }

  // At this point, we have a bunch of data. Replace the contents of
  // event-list with our new data.
  $('#event-list').html(''); // clear list
  $('#event-list').append(frag);

}

function initializeApp() {
  $.ajax({
    url: "/event/list",
    method: "POST",
    success: function(data) {
      var parsed = JSON.parse(data);
      if (parsed.status === "error") {
        flashMessage("Error: " + parsed.message);
        return;
      }
      // status === "success"

      // The data must be a list of events b/c it was successful.
      listEvents(parsed);
    },
    error: function() {
      flashMessage("Error connecting to server.");
    },
  });
}
