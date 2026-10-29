// Sign-in page for members.directactioneverywhere.com. Loaded as a separate
// file (the page's CSP forbids inline script). Configuration arrives via data-*
// attributes on #login.
(function () {
  "use strict";

  var root = document.getElementById("login");
  var clientId = root.dataset.clientId;
  var nonce = root.dataset.nonce;
  var next = root.dataset.next;
  var dark = window.matchMedia("(prefers-color-scheme: dark)");

  function $(id) {
    return document.getElementById(id);
  }
  function show(el, text) {
    el.textContent = text;
    el.hidden = false;
  }
  function hide(el) {
    el.hidden = true;
  }

  function post(path, body) {
    return fetch(path, {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  }

  function signedIn() {
    window.location.replace(next);
  }

  // --- Google ---

  var googleError = $("google-error");

  function onCredential(response) {
    hide(googleError);
    post("/auth/google/token", { credential: response.credential })
      .then(function (res) {
        if (res.ok) return signedIn();
        if (res.status === 429) {
          show(googleError, "Too many attempts. Please wait a moment and try again.");
        } else if (res.status === 401) {
          show(googleError, "Google sign-in failed or expired. Reload the page and try again, or use email.");
        } else {
          show(googleError, "Something went wrong signing in with Google. Please try again, or use email.");
        }
      })
      .catch(function () {
        show(googleError, "Couldn't reach the server. Check your connection and try again.");
      });
  }

  function renderGoogleButton() {
    var el = $("google-button");
    el.textContent = "";
    var width = Math.max(200, Math.min(400, Math.floor(el.clientWidth || 320)));
    window.google.accounts.id.renderButton(el, {
      type: "standard",
      theme: dark.matches ? "filled_black" : "outline",
      size: "large",
      text: "continue_with",
      shape: "rectangular",
      logo_alignment: "left",
      width: width,
    });
  }

  function initGoogle() {
    if (!clientId) {
      show(googleError, "Google sign-in isn't available right now. Please use email.");
      return;
    }
    window.google.accounts.id.initialize({
      client_id: clientId,
      callback: onCredential,
      nonce: nonce,
      auto_select: false,
      ux_mode: "popup",
    });
    renderGoogleButton();
    if (dark.addEventListener) dark.addEventListener("change", renderGoogleButton);
  }

  function loadGoogle() {
    var failed = function () {
      show(googleError, "Google sign-in couldn't load (it may be blocked by your browser or an extension). You can use email instead.");
    };
    var s = document.createElement("script");
    s.src = "https://accounts.google.com/gsi/client";
    s.async = true;
    s.onload = function () {
      if (window.google && window.google.accounts) initGoogle();
      else failed();
    };
    s.onerror = failed;
    document.head.appendChild(s);
  }

  loadGoogle();

  // --- Email link ---

  var emailForm = $("email-form");
  var emailInput = $("email");
  var emailSubmit = $("email-submit");
  var emailStatus = $("email-status");
  var emailError = $("email-error");

  emailForm.addEventListener("submit", function (e) {
    e.preventDefault();
    hide(emailStatus);
    hide(emailError);
    var email = emailInput.value.trim();
    if (!email || !emailInput.checkValidity()) {
      show(emailError, "Enter a valid email address.");
      emailInput.focus();
      return;
    }
    emailSubmit.disabled = true;
    post("/auth/email", { email: email, next: next })
      .then(function (res) {
        if (res.ok) {
          show(emailStatus, "Check your email for a sign-in link.");
        } else if (res.status === 429) {
          show(emailError, "Too many requests. Please wait a while and try again.");
        } else if (res.status === 400) {
          show(emailError, "That doesn't look like a valid email address.");
        } else {
          show(emailError, "We couldn't send the link. Please try again.");
        }
      })
      .catch(function () {
        show(emailError, "Couldn't reach the server. Check your connection and try again.");
      })
      .then(function () {
        emailSubmit.disabled = false;
      });
  });

  // --- Dev ---

  var devButton = $("dev-signin");
  if (devButton) {
    devButton.addEventListener("click", function () {
      post("/auth/google/token", { credential: "dev" }).then(function (res) {
        if (res.ok) signedIn();
        else show(googleError, "Dev sign-in failed.");
      });
    });
  }
})();
