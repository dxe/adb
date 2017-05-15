import ActivistList from 'ActivistList.vue';
import Vue from 'vue';
//import App from 'App.vue';

var vm = new Vue({
  el: "#app",
  data: {
    activists: [
      {
        name: "Bob Jones",
        email: "bob@bob.com",
        phone: "703-225-8132",
        first_event: "01-01-2017",
        last_event: "02-01-2017",
        status: "New",
      },
    ],
  },
  render: function(h) {
    return h(ActivistList, {activists: this.activists});
  }
});





// var ACTIVISTS = [];

// var DAYS_AGO_60 = new Date();
// DAYS_AGO_60.setDate(DAYS_AGO_60.getDate() - 60);

// var DAYS_AGO_90 = new Date();
// DAYS_AGO_90.setDate(DAYS_AGO_90.getDate() - 90);

// function getActivistStatusHTML(activist, firstEventDate, lastEventDate) {
//   if (activist.name == "") {
//     return "";
//   }

//   if (!firstEventDate || !lastEventDate) {
//     return '<span class="hide">4</span>No attendance';
//   }

//   if (lastEventDate < DAYS_AGO_60) {
//     return '<span class="hide">3</span>Former';
//   } else if (firstEventDate > DAYS_AGO_90  && activist.total_events < 5) {
//     return '<span class="hide">2</span>New';
//   }

//   return '<span class="hide">1</span>Current';
// }

// function listActivists(activists) {
//   $("#activist-list-body").html('');
//   var d = document.getElementById('activist-list-body');
//   var m = document.getElementById('modals');

//   for (var i = 0; i < activists.length; i++) {
//     var activist = activists[i];

//     var firstEventDate = null;
//     var lastEventDate = null;

//     if (activist.first_event !== "") {
//       var firstEventSplit = activist.first_event.split('-');
//       firstEventDate = new Date(firstEventSplit[0],firstEventSplit[1]-1,firstEventSplit[2]);
//     }

//     if (activist.last_event !== "") {
//       var lastEventSplit = activist.last_event.split('-');
//       lastEventDate = new Date(lastEventSplit[0],lastEventSplit[1]-1,lastEventSplit[2]);
//     }

//     var activistStatus = getActivistStatusHTML(activist, firstEventDate, lastEventDate);

//     var modalID = 'modal' + activist.id;
//     var modal = '<div class="modal fade" id=' + modalID + ' tabindex="-1" role="dialog">' +
//         '<div class="modal-dialog" role="document"><div class="modal-content"><div class="modal-header">' +
//         '<h2 class="modal-title">' + activist.name + '</h5></div>' +
//         '<div class="modal-body">' +
//         '<label for="email">Email: </label><input class="form-control" type="text" value="'+ activist.email + '" id="email"><br />' +
//         '<label for="chapter">Chapter: </label><input class="form-control" type="text" value="'+ activist.chapter + '" id="chapter"><br />' +
//         '<label for="phone">Phone: </label><input class="form-control" type="text" value="'+ activist.phone + '" id="phone"><br />' +
//         '<label for="location">Location: </label><input class="form-control" type="text" value="'+ activist.location + '" id="location"><br />' +
//         '<label for="facebook">Facebook: </label><input class="form-control" type="text" value="'+ activist.facebook + '" id="facebook"><br />' +
//         '<label for="core">Core/Staff:&nbsp;</label><input class="form-check-input" type="checkbox" id="core"><br />' +
//         '<label for="exclude">Exclude from Leaderboard:&nbsp;</label><input class="form-check-input" type="checkbox" id="exclude"><br />' +
//         '<label for="pledge">Liberation Pledge:&nbsp;</label><input class="form-check-input" type="checkbox" id="pledge"><br />' +
//         '<label for="globalteam">Global Team Member:&nbsp;</label><input class="form-check-input" type="checkbox" id="globalteam">' +
//         '<div class="modal-footer"><button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button><button type="button" class="btn btn-primary">Save changes</button>' +
//         '</div></div></div></div>'

//     var newRow = '<tr>' +
//         '<td>' + '<a class="edit-link" href="#" data-toggle="modal" data-target="#' + modalID + '"><button class="btn btn-default glyphicon glyphicon-pencil"></button></a>' + '</td>' +
//         '<td>' + activist.name + '</td>' +
//         '<td>' + activist.email + '</td>' +
//         '<td>' + activist.phone + '</td>' +
//         '<td>' + activist.first_event + '</td>' +
//         '<td>' + activist.last_event + '</td>' +
//         '<td>' + activistStatus + '</td>' +
//         '</tr>';
//     d.insertAdjacentHTML('beforeend', newRow);
//     m.insertAdjacentHTML('beforeend', modal);
//   }
//   $(function(){ $("#activist-list").tablesorter({ sortList: [[6,0], [1,0]] }); });
// }

// export function initializeApp() {
//   $.ajax({
//     url: "/activist/list",
//     success: function(data) {
//       var parsed = JSON.parse(data);
//       if (parsed.status === "error") {
//         flashMessage("Error: " + parsed.message, true);
//         return;
//       }
//       // status === "success"

//       ACTIVISTS = parsed;
//       listActivists(parsed);
//     },
//     error: function() {
//       flashMessage("Error connecting to server.", true);
//     },
//   });
// }
