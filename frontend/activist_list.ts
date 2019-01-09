import ActivistList from './ActivistList.vue';
import Vue from 'vue';

export function initializeApp(view) {
  var vm = new Vue({
    el: '#app',
    render(h) {
      return h(ActivistList, {
        props: {
          view: view,
        },
      });
    },
  });
}
