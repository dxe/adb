function listEvents(events) {
  if (events.length === 0) {
    flashMessage("No events from server", true);
    return;
  }

  // First, clear body.
  $('#event-list-body').html('');

  for (var i = 0; i < events.length; i++) {
    var event = events[i];
    var attendeeString = '';
    for (var j = 0; j < event.attendees.length; j++) {
      attendeeString += '<li>' + event.attendees[j]; '</li>';
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
        '<td><ul class="attendee-list">' + attendeeString + '</ul></td>' +
        '</tr>';
    var d = document.getElementById('event-list-body');
    d.insertAdjacentHTML('beforeend', newRow);
  }
}

function eventListRequest() {
  var eventDateStart = $('#event-date-start').val();
  var eventDateEnd = $('#event-date-end').val();

  $.ajax({
    url: "/event/list",
    method: "POST",
    data: {
      event_date_start: eventDateStart,
      event_date_end: eventDateEnd,
    },
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

function initDateRange() {
  // First, set event-date-start
  var d = new Date();
  var rawYear = d.getFullYear();
  var rawMonth = d.getMonth() + 1;

  // Set the "from" date to the 1st of last month.
  if (rawMonth == 1) {
    rawMonth = 12;
    rawYear -= 1;
  } else {
    rawMonth -= 1;
  }

  var year = '' + rawYear;
  var month = (rawMonth > 9) ? '' + rawMonth : '0' + rawMonth;

  var fromDate = year + '-' + month + '-01';
  $('#event-date-start').val(fromDate);
  
  // set "to" date to today
  var toDate = d.toISOString().slice(0, 10);
  $('#event-date-end').val(toDate);

}

function initializeApp() {
  initDateRange();
  eventListRequest();
}
