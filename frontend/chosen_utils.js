import 'bootstrap-chosen/bootstrap-chosen.css';
import 'chosen-js'; // Attaches to jQuery when it's imported.
import { flashMessage } from './flash_message';

// From chosen-js
function chosenBrowserIsSupported() {
  if ('Microsoft Internet Explorer' === window.navigator.appName) {
    return document.documentMode >= 8;
  }
  if (
    /iP(od|hone)/i.test(window.navigator.userAgent) ||
    /IEMobile/i.test(window.navigator.userAgent) ||
    /Windows Phone/i.test(window.navigator.userAgent) ||
    /BlackBerry/i.test(window.navigator.userAgent) ||
    /BB10/i.test(window.navigator.userAgent) ||
    /Android.*Mobile/i.test(window.navigator.userAgent)
  ) {
    return false;
  }
  return true;
}

export function initActivistSelect(selector, ignoreActivistName) {
  var $selector = $(selector);

  // Chosen-js isn't supported on mobile browsers. We need to add the
  // class "form-control" to the selector if it isn't supported so the
  // selector doesn't look super ugly.
  if (!chosenBrowserIsSupported()) {
    $selector.addClass('form-control');
  }

  $.ajax({
    url: '/activist_names/get',
    method: 'GET',
    dataType: 'json',
    success: () => {
      var activistNames = data.activist_names;

      // The first item needs to be empty so that the selector
      // defaults to not being filled in. Otherwise, it defaults to
      // the first item in the list.
      activistNames.unshift('');

      for (var i = 0; i < activistNames.length; i++) {
        if (activistNames[i] == ignoreActivistName) {
          continue;
        }

        $selector[0].options.add(new Option(activistNames[i]));
      }

      $selector.chosen({
        allow_single_deselect: true,
        inherit_select_classes: true,
      });
    },
    error: () => {
      flashMessage('Error: could not load activist names', true);
    },
  });
}
