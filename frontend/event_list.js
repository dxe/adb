import {flashMessage} from 'flash_message';
import 'bootstrap-chosen/bootstrap-chosen.css';
import 'chosen-js'; // Attaches to jQuery when it's imported.

// From chosen-js
function chosenBrowserIsSupported() {
  if ("Microsoft Internet Explorer" === window.navigator.appName) {
    return document.documentMode >= 8;
  }
  if (/iP(od|hone)/i.test(window.navigator.userAgent) ||
      /IEMobile/i.test(window.navigator.userAgent) ||
      /Windows Phone/i.test(window.navigator.userAgent) ||
      /BlackBerry/i.test(window.navigator.userAgent) ||
      /BB10/i.test(window.navigator.userAgent) ||
      /Android.*Mobile/i.test(window.navigator.userAgent)) {
    return false;
  }
  return true;
}

export function confirmDeleteEvent(eventID) {
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
        '<div class="dropdown"><button class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal" data-toggle="dropdown"></button><ul class="dropdown-menu"><li><a href="javascript:event_list.confirmDeleteEvent(' + event.event_id + ')">Delete event</a></li></ul></div>' +
        '</td>' +
        '<td nowrap>' + event.event_date + '</td>' +
        '<td class="event-name"><b>' + event.event_name + '</b></td>' +
        '<td nowrap>' + event.event_type + '</td>' +
        '<td nowrap>' + event.attendees.length + '</td>' +
        '<td>' +
          '<button class="show-attendees btn btn-link" onclick="event_list.toggleAttendees(\'' + rowID + '\')" >+ Attendees</button>' +
          '<ul class="attendee-list" style="display: none">' + attendeeString + '</ul></td>' +
        '</tr>';
    d.insertAdjacentHTML('beforeend', newRow);
  }
}

export function toggleAttendees(rowID) {
  var $row = $('#' + rowID);
  var $showAttendeesBtn = $row.find('.show-attendees');
  var $attendees = $row.find('.attendee-list');
  if ($showAttendeesBtn.text() === "+ Attendees") {
    $attendees.show();
    $showAttendeesBtn.text('- Attendees');
  } else {
    $attendees.hide();
    $showAttendeesBtn.text('+ Attendees');
  }
}

export function showAllAttendees() {
  $('.show-attendees').text('- Attendees');
  $('.attendee-list').show();
}

export function hideAllAttendees() {
  $('.show-attendees').text('+ Attendees');
  $('.attendee-list').hide();
}

export function eventListRequest() {
  var eventName = $('#event-name').val();
  var eventActivist = $('#event-activist').val();
  var eventDateStart = $('#event-date-start').val();
  var eventDateEnd = $('#event-date-end').val();
  var eventType = $('#event-type').val();

  $.ajax({
    url: "/event/list",
    method: "POST",
    data: {
      event_name: eventName,
      event_activist: eventActivist,
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

function initEventActivistSelect() {
  var $selector = $("#event-activist");

  // Chosen-js isn't supported on mobile browsers. We need to add the
  // class "form-control" to the selector if it isn't supported so the
  // selector doesn't look super ugly.
  if (!chosenBrowserIsSupported()) {
    $selector.addClass('form-control');
  }


  $.ajax({
    url: "/activist_names/get",
    method: "GET",
    dataType: "json",
    success: function(data) {
      var activistNames = data.activist_names;

      activistNames.unshift("");

      for (var i = 0; i < activistNames.length; i++) {
        $selector[0].options.add(new Option(activistNames[i]));
      }

      $selector.chosen({
        allow_single_deselect: true,
        inherit_select_classes: true,
      });
    },
    error: function() {
      flashMessage("Error: could not load activist names", true);
    },
  });
}

export function initializeApp() {
  initDateRange();
  initEventActivistSelect();
  eventListRequest();
}
