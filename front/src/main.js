import Vue from 'vue'
import App from './App.vue'
import './plugins/element.js'
import axios from 'axios'
import { backUrl } from '@/config/index.js'

Vue.config.productionTip = false

Vue.prototype.$http = axios

axios.defaults.baseURL = backUrl;

new Vue({
  render: h => h(App),
}).$mount('#app')
