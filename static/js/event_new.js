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
var ACTIVIST_NAMES_SET = new Set();

function updateAutocompleteNames() {
  $.ajax({
    url: "/activist_names/get",
    method: "GET",
    dataType: "json",
    success: function(data) {
      var activistNames = data.activist_names;
      for (var i = 0; i < activistNames.length; i++) {
        ACTIVIST_NAMES.push(activistNames[i]);
        ACTIVIST_NAMES_SET.add(activistNames[i]);
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
  //
  // Change fires if the form is changed and the user moves onto the
  // next input (e.g. the user types in a name and presses tab).
  $('#eventForm').change(function(e) {
    DIRTY = true;
  });
  // Input is fired any time the user types in an input field.
  $('#attendee-rows').on('input', function(e) {
    DIRTY = true;
    var input = e.target;
    updateInputColor(input);
    maybeExpandRows(input);
  });
  // awesomplete-selectcomplete is fired when the user clicks on a
  // name in the awesomplete dropdown.
  $('#attendee-rows').on("awesomplete-selectcomplete", function(e) {
    DIRTY = true;
    var input = e.target;
    updateInputColor(input);
    maybeExpandRows(input);

    // Select the next row.
    var $rows = $('.attendee-input');
    for (var i = 0; i < $rows.length; i++) {
      var row = $rows[i];
      if (input === row) {
        // Select the row after.
        if (i+1 < $rows.length) {
          $($rows[i+1]).focus();
        }
      }
    }
  });
}

// Update the color of the input element.
// Highlight in red if the input is a duplicate.
// Highlight in yellow if the user is not in the database.
function updateInputColor(input) {
  var value = input.value;
  // If the input is blank, just remove the style.
  if (value === '') {
    input.style.border = '';
    return;
  }

  var theEntireRows = document.querySelector('#attendee-rows');
  var currentValues = new Set();
  for (var i = 0; i< theEntireRows.children.length; i++) {
    // insert the values into the Set only if it not null
    if (input !== theEntireRows.children[i].children[0] && theEntireRows.children[i].children[0].value !== "") {
      currentValues.add(theEntireRows.children[i].children[0].value)
    }
  }

  if (currentValues.has(value)) {
    // If the name is a duplicate of all the names entered, color it
    // red.
    input.style.border = '2px solid red';
  } else if (!ACTIVIST_NAMES_SET.has(value)) {
    // If the name is not in the set of all activist names, then color it yellow.
    input.style.border = '2px solid yellow';
  } else {
    input.style.border = '';
  }
}

// Expand the number of rows automatically if one of the last two
// rows has a value.
function maybeExpandRows(currentInput) {
  var $rows = $('.attendee-input');
  if ($rows[$rows.length - 1].value !== '' ||
      $rows[$rows.length - 2].value !== '') {
    addRows(5);
  }

  // After expanding, focus on the current input again.
  $(currentInput).focus();
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
