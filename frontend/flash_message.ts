import { getCookie, deleteCookie } from './cookie';

import { ToastProgrammatic as Toast } from 'buefy';

export function flashMessage(content: string, isError?: boolean) {
  const toastType = isError ? 'is-danger' : 'is-success';
  const duration = 3000;
  const position = 'is-bottom-right';
  Toast.open({ type: toastType, position: position, duration: duration, message: content });
}

export function initializeFlashMessage() {
  // First, get any potential messages. Will be undefined if the
  // cookie isn't set.
  const error = getCookie('flash_message_error');
  const success = getCookie('flash_message_success');

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

export function setFlashMessageSuccessCookie(content: string) {
  document.cookie = 'flash_message_success=' + encodeURIComponent(content) + ';path=/';
}

export function setFlashMessageErrorCookie(content: string) {
  document.cookie = 'flash_message_error=' + encodeURIComponent(content) + ';path=/';
}
