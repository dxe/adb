// DIRTY represents whether the form has been modified before the user
// has saved. It is set to false when the user saves.
var DIRTY = false;

window.addEventListener('beforeunload', function(e) {
  if (!DIRTY) {
    return;
  }
  var message = "You have unsaved data.";
  e.returnValue = message;
  return message;
});

var ACTIVIST_NAMES = [];

function updateAutocompleteNames() {
  $.ajax({
    url: "/activist_names/get",
    method: "GET",
    dataType: "json",
    success: function(data) {
      var activistNames = data.activist_names;
      for (var i = 0; i < activistNames.length; i++) {
        ACTIVIST_NAMES.push(activistNames[i]);
      }
    },
    error: function() {
      flashMessage("Error: could not load activist names", true);
    },
  });
}

function updateAwesomeplete() {
  var $attendeeRows = $('.attendee-input');

  for (var i = 0; i < $attendeeRows.length; i++) {
    new Awesomplete($attendeeRows[i], { list: ACTIVIST_NAMES });
  }
}

function initializeApp() {
  addRows(5);
  updateAutocompleteNames();
  // If any form input/selection changes, mark the page as dirty.
  $('#eventForm').change(function(e) {
    DIRTY = true;
  });
}

// creates new event in ADB
function newEvent(event) {
  var eventName = document.getElementById('eventName').value;
  if (eventName === "") {
    flashMessage("Error: Please enter event name!", true);
    return;
  }

  var eventDate = document.getElementById('eventDate').value;
  if (eventDate === "") {
    flashMessage("Error: Please enter date!", true);
    return;
  }

  var eventType = document.getElementById('eventType').value;
  if (eventType === "") {
    flashMessage("Error: Must choose event type.", true);
    return;
  }

  var attendees = [];
  var $attendeeRows = $('.attendee-input');
  for (var i = 0; i < $attendeeRows.length; i++) {
    var attendeeValue = $attendeeRows[i].value;
    if (attendeeValue !== "") {
      attendees.push(attendeeValue);
    }
  }

  if (attendees.length === 0) {
    flashMessage("Error: must enter attendees", true);
    return;
  }

  var eventID = parseInt(document.getElementById('eventID').value);

  $.ajax({
    url: "/event/save",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      event_id: eventID,
      event_name: eventName,
      event_date: eventDate,
      event_type: eventType,
      attendees: attendees,
    }),
    success: function(data) {
      var parsed = JSON.parse(data);
      if (parsed.status === "error") {
        flashMessage("Error: " + parsed.message, true);
        return;
      }
      // status === "success"
      // Saved successfully, mark the page as clean.
      DIRTY = false;

      if (parsed.redirect) {
        setFlashMessageSuccessCookie("Saved!");
        window.location = parsed.redirect;
      } else {
        flashMessage("Saved!", false);
      }
    },
    error: function() {
      flashMessage("Error, did not save data", true);
    },
  });
}

function addRows(numToAdd) {
  var $rowsContainer = $('#attendee-rows');

  for (var i = 0; i < numToAdd; i++) {
    $rowsContainer.append("<input class='attendee-input form-control' />");
  }

  updateAwesomeplete();
}

function setDateToToday(event) {
  var d = new Date();
  var year = '' + d.getFullYear();

  var rawMonth = d.getMonth() + 1;
  var month = (rawMonth > 9) ? '' + rawMonth : '0' + rawMonth;

  var rawDate = d.getDate();
  var date = (rawDate > 9) ? '' + rawDate : '0' + rawDate;

  var validDateString = year + '-' + month + '-' + date;

  $("#eventDate").val(validDateString);
}
