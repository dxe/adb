import Vue from 'vue';
import * as Sentry from '@sentry/browser';
import * as Integrations from '@sentry/integrations';
import ActivistList from './ActivistList.vue';
import CirclesList from './CirclesList.vue';
import EventEdit from './EventEdit.vue';
import EventList from './EventList.vue';
import UserList from './UserList.vue';
import WorkingGroupList from './WorkingGroupList.vue';

Sentry.init({
  dsn: 'https://1bc5adff2f574d5390f085353326f0d5@sentry.io/1820807',
  integrations: [new Integrations.Vue({ Vue, attachProps: true })],
});

new Vue({
  el: '#app',
  components: {
    ActivistList,
    CirclesList,
    EventEdit,
    EventList,
    UserList,
    WorkingGroupList,
  },
});
