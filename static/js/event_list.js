function confirmDeleteEvent(eventID) {
  var eventRow = $("#event-id-" + eventID);
  var eventName = eventRow.find(".event-name").text();
  var confirmed = confirm('Are you sure you want to delete the event "' + eventName + '"?\n\nPress OK to delete this event.');

  if (confirmed) {
    $.ajax({
      url: "/event/delete",
      method: "POST",
      data: {
        event_id: eventID,
      },
      success: function(data) {
        var parsed = JSON.parse(data);
        if (parsed.status === "error") {
          flashMessage("Error: " + parsed.message, true);
          return;
        }
        // status === "success"

        flashMessage("Deleted event " + eventName);
        eventListRequest();
      },
      error: function() {
        flashMessage("Error connecting to server.", true);
      },
    });
  }

  return false;
}

function listEvents(events) {
  if (events.length === 0) {
    flashMessage("No events from server", true);
    return;
  }

  // First, clear body.
  $('#event-list-body').html('');

  var d = document.getElementById('event-list-body');

  for (var i = 0; i < events.length; i++) {
    var event = events[i];
    var attendeeString = '';
    for (var j = 0; j < event.attendees.length; j++) {
      attendeeString += '<li>' + event.attendees[j]; '</li>';
    }

    // Now, create the links.
    var updateLink = '/update_event/' + event.event_id;

    var rowID = "event-id-" + event.event_id;

    // output to new row in table to display
    var newRow = '<tr id="' + rowID + '">' +
        '<td>' +
        '<a class="edit-link" href="' + updateLink + '"><button class="btn btn-default glyphicon glyphicon-pencil"></button></a><br/><br/>' +
        '<div class="dropdown"><button class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal" data-toggle="dropdown"></button><ul class="dropdown-menu"><li><a href="javascript:confirmDeleteEvent(' + event.event_id + ')">Delete event</a></li></ul></div>' +
        '</td>' +
        '<td nowrap>' + event.event_date + '</td>' +
        '<td class="event-name"><b>' + event.event_name + '</b></td>' +
        '<td nowrap>' + event.event_type + '</td>' +
        '<td nowrap>' + event.attendees.length + '</td>' +
        '<td><ul class="attendee-list">' + attendeeString + '</ul></td>' +
        '</tr>';
    d.insertAdjacentHTML('beforeend', newRow);
  }
}

function eventListRequest() {
  var eventDateStart = $('#event-date-start').val();
  var eventDateEnd = $('#event-date-end').val();
  var eventType = $('#event-type').val();

  $.ajax({
    url: "/event/list",
    method: "POST",
    data: {
      event_date_start: eventDateStart,
      event_date_end: eventDateEnd,
      event_type: eventType,
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
