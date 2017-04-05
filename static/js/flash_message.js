function flashMessage(content) {
  var flash = $('#flash');
  flash.text(content);
  flash.show();

  setTimeout(function() {
    flash.hide();
  }, 60 * 5);
}
