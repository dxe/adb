// Called by the google sign in button.
function onSignIn(googleUser) {
  var id_token = googleUser.getAuthResponse().id_token;
  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/tokensignin');
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.onload = function() {
    if (xhr.status !== 200) {
      $("#message").text("Error status code from server: " + xhr.status)
      return;
    }

    try {
      var data = JSON.parse(xhr.responseText);
    } catch(err) {
      $('#message').text("Could not parse message from server: " + xhr.responseText);
      return;
    }

    if (data['redirect']) {
      // Redirect to index
      window.location = "/";
      return;
    } else {
      // Print the message if it exists
      if (data['message']) {
        $('#message').text(data['message']);
      } else {
        $('#message').text("Server could not authenticate you for some reason");
      }
    }
  };
  xhr.send('idtoken=' + id_token);
}
