import * as Awesomplete from 'external/awesomplete';
import {flashMessage, setFlashMessageSuccessCookie} from 'flash_message';

// DIRTY represents whether the form has been modified before the user
// has saved. It is set to false when the user saves.
var DIRTY = false;

function initializeDirty() {
  window.addEventListener('beforeunload', function(e) {
    if (!DIRTY) {
      return;
    }
    var message = "You have unsaved data.";
    e.returnValue = message;
    return message;
  });
}

/* All activists from database */
var ACTIVIST_NAMES = [];
var ACTIVIST_NAMES_SET = new Set();

/* Activists associated just with this event */
var EVENT_ATTENDEE_NAMES = [];
var EVENT_ATTENDEE_NAMES_SET = new Set();

function updateAutocompleteNames() {
  $.ajax({
    url: "/activist_names/get",
    method: "GET",
    dataType: "json",
    success: function(data) {
      var activistNames = data.activist_names;
      // Clear current activist name array and set before re-adding
      ACTIVIST_NAMES.length = 0;
      ACTIVIST_NAMES_SET.clear();
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

/* Retrieve attendance before any edits are made */
/* Should I safeguard against any malformed data? */
function getEventAttendeeNames(eventAttendees) {
    if (eventAttendees === null) {
        // No existing data. Must be a new event
        return;
    }
    EVENT_ATTENDEE_NAMES = eventAttendees.map(function(attendee) {
        EVENT_ATTENDEE_NAMES_SET.add(attendee.Name);
        return attendee.Name; 
    });
}

function updateAwesomeplete() {
  // Only grab inputs that are not children of div.awesomplete
  // Note length of $attendeeRows = 0 if there are no input.attendee-input elements
  var $attendeeRows = $('#attendee-rows > input.attendee-input');
    
  for (var i = 0; i < $attendeeRows.length; i++) {
    new Awesomplete($attendeeRows[i], { list: ACTIVIST_NAMES });
  }
}

export function initializeApp(eventAttendees) {
  initializeDirty();
  addRows(5);
  getEventAttendeeNames(eventAttendees);
  updateAutocompleteNames();
  countAttendees();
  // If any form input/selection changes, mark the page as dirty.
  //
  // Change fires if the form is changed and the user moves onto the
  // next input (e.g. the user types in a name and presses tab).
  $('#eventForm').change(function(e) {
    DIRTY = true;
  });
  initAttendeeInputEventHandlers();
}

function initAttendeeInputEventHandlers() {
  // Input is fired any time the user types in an input field.
  $('#attendee-rows').on('input', function(e) {
    DIRTY = true;
    var input = e.target;
    updateInputColor(input);
    maybeExpandRows(input);
    countAttendees(input);
  });
  // awesomplete-selectcomplete is fired when the user clicks on a
  // name in the awesomplete dropdown.
  $('#attendee-rows').on("awesomplete-selectcomplete", function(e) {
    DIRTY = true;
    var input = e.target;
    updateInputColor(input);
    maybeExpandRows(input);
    countAttendees(input);

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
      currentValues.add(theEntireRows.children[i].children[0].value);
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
    addRows(1);
  }

  // After expanding, focus on the current input again.
  if (typeof currentInput !== 'undefined') {
    $(currentInput).focus();
  }
}

function countAttendees(currentInput) {
  var $rows = $('.attendee-input');
  var attendeeTotal = 0;
  for (var i = 0; i < $rows.length; i++) {
    if ($rows[i].value !== '') {
      attendeeTotal += 1;
    }
  }
  $('#attendeeTotal').html(attendeeTotal);
}

// creates new event in ADB
export function newEvent(event) {
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
  var attendeesSet = new Set();
  var $attendeeRows = $('.attendee-input');
  for (var i = 0; i < $attendeeRows.length; i++) {
    var attendeeValue = $attendeeRows[i].value.trim();
    if (attendeeValue !== "") {
      attendees.push(attendeeValue);
      attendeesSet.add(attendeeValue);
    }
  }

  if (attendees.length === 0) {
    flashMessage("Error: must enter attendees", true);
    return;
  }

  var eventID = parseInt(document.getElementById('eventID').value);
  var addedActivists = attendees.filter(function (activist) {
        return !EVENT_ATTENDEE_NAMES_SET.has(activist);
  });
  var deletedActivists = EVENT_ATTENDEE_NAMES.filter(function (activist) {
      return !attendeesSet.has(activist);
  });

  $.ajax({
    url: "/event/save",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify({
      event_id: eventID,
      event_name: eventName,
      event_date: eventDate,
      event_type: eventType,
      added_attendees: addedActivists,
      deleted_attendees: deletedActivists,
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
        refreshEventAttendance(parsed.attendees)
      }
    },
    error: function() {
      flashMessage("Error, did not save data", true);
    },
  });

}

function refreshEventAttendance(attendees) {
    var numberOfAttendees = 0;
    if (attendees === null) {
        // All attendees were deleted from this event
        EVENT_ATTENDEE_NAMES = [];
        EVENT_ATTENDEE_NAMES_SET = new Set();
    }
    else {
        EVENT_ATTENDEE_NAMES = attendees;
        EVENT_ATTENDEE_NAMES_SET = new Set(EVENT_ATTENDEE_NAMES);
        numberOfAttendees = attendees.length;
    }

    $('#attendee-rows').empty(); // clear existing html
    addRows(numberOfAttendees);
    var attendeeList = $('#attendee-rows').find('.attendee-input');
    for (var i = 0; i < numberOfAttendees; i++) {
        attendeeList[i].value = EVENT_ATTENDEE_NAMES[i];
    }
    $('#attendeeTotal').html(numberOfAttendees); // update total attendee counter
    addRows(5);
    updateAutocompleteNames();
    initAttendeeInputEventHandlers();
}

function addRows(numToAdd) {
  var $rowsContainer = $('#attendee-rows');

  for (var i = 0; i < numToAdd; i++) {
    $rowsContainer.append("<input class='attendee-input form-control' />");
  }

  updateAwesomeplete();
}

export function setDateToToday(event) {
  var d = new Date();
  var year = '' + d.getFullYear();

  var rawMonth = d.getMonth() + 1;
  var month = (rawMonth > 9) ? '' + rawMonth : '0' + rawMonth;

  var rawDate = d.getDate();
  var date = (rawDate > 9) ? '' + rawDate : '0' + rawDate;

  var validDateString = year + '-' + month + '-' + date;

  $("#eventDate").val(validDateString);
}
