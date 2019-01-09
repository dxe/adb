import { getCookie, deleteCookie } from './cookie';

export function flashMessage(content, isError) {
  var flash = $('#flash');
  if (isError) {
    flash[0].className = 'alert alert-danger';
  } else {
    flash[0].className = 'alert alert-success';
  }
  flash.text(content);
  flash.show();

  setTimeout(function() {
    flash.hide();
  }, 5 * 1000);
}

export function initializeFlashMessage() {
  // First, get any potential messages. Will be undefined if the
  // cookie isn't set.
  var error = getCookie('flash_message_error');
  var success = getCookie('flash_message_success');

  // Clear cookies that exist
  if (error) {
    deleteCookie('flash_message_error');
  }
  if (success) {
    deleteCookie('flash_message_success');
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

export function setFlashMessageSuccessCookie(content) {
  document.cookie = 'flash_message_success=' + encodeURIComponent(content) + ';path=/';
}

export function setFlashMessageErrorCookie(content) {
  document.cookie = 'flash_message_error=' + encodeURIComponent(content) + ';path=/';
}
