import UserList from 'UserList.vue';
import Vue from 'vue';

export function initializeApp() {
  var vm = new Vue({
    el: '#app',
    render(h) {
      return h(UserList);
    },
  });
}
