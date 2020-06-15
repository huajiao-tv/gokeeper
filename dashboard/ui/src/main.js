import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import router from './router/index'
import './api/request'

import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import '@/style/index.scss' // global css

Vue.use(ElementUI)

// import './mock/index'

Vue.config.productionTip = false

new Vue({
  vuetify,
  router,
  render: h => h(App)
}).$mount('#app')
