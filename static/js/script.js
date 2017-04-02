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
    // TODO: handle errors
  });
}

function updateAwesomeplete() {
  var $attendeeRows = $('.attendee-input');

  for (var i = 0; i < $attendeeRows.length; i++) {
    new Awesomplete($attendeeRows[i], { list: ACTIVIST_NAMES });
  }
}

function initializeApp() {
  addRows(10);
  updateAutocompleteNames();
}

// creates new event in ADB
function newEvent(event) {
  // TODO: use modals
  var eventName = document.getElementById('eventName').value;
  if (eventName === "") {
    alert("Error: Please enter event name!");
    return;
  }

  var eventDate = document.getElementById('eventDate').value;
  if (eventDate == "") {
    alert("Error: Please enter date!");
    return;
  }

  var eventType = document.getElementById('eventType').value;

  var attendees = [];
  var $attendeeRows = $('.attendee-input');
  for (var i = 0; i < $attendeeRows.length; i++) {
    var attendeeValue = $attendeeRows[i].value;
    if (attendeeValue !== "") {
      attendees.push(attendeeValue);
    }
  }

  if (attendees.length === 0) {
    alert("Error: must enter attendees");
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
    // TODO: deal with success and error
  });
}

function addRows(numToAdd) {
  var $rowsContainer = $('#attendee-rows');

  for (var i = 0; i < numToAdd; i++) {
    $rowsContainer.append("<input class='attendee-input' />");
  }

  updateAwesomeplete();
}
