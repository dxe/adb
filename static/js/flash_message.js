function flashMessage(content, isError) {
  var flash = $('#flash');
  if (isError) {
    flash[0].className = "alert alert-danger";
  } else {
    flash[0].className = "alert alert-success";
  }
  flash.text(content);
  flash.show();

  setTimeout(function() {
    flash.hide();
  }, 5 * 1000);
}

// from http://stackoverflow.com/questions/10730362/get-cookie-by-name
function getCookie(name) {
  var value = "; " + document.cookie;
  var parts = value.split("; " + name + "=");
  if (parts.length == 2) return parts.pop().split(";").shift();
}

// from http://stackoverflow.com/questions/2144386/javascript-delete-cookie
function deleteCookie(name) {
  document.cookie = name + '=;path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

function initializeFlashMessage() {
  // First, get any potential messages. Will be undefined if the
  // cookie isn't set.
  var error = getCookie("flash_message_error");
  var success = getCookie("flash_message_success");

  // Clear cookies that exist
  if (error) {
    deleteCookie("flash_message_error");
  }
  if (success) {
    deleteCookie("flash_message_success");
  }

  // Show the error message if it exists. Otherwise, show the success
  // message.
  if (error) {
    flashMessage(error, true);
    return;
  }
  if (success) {
    flashMessage(success, false);
    return;
  }
}

function setFlashMessageSuccessCookie(content) {
  document.cookie = "flash_message_success=" + encodeURIComponent(content) + ";path=/";
}

function setFlashMessageErrorCookie(content) {
  document.cookie = "flash_message_error=" + encodeURIComponent(content) + ";path=/";
}
