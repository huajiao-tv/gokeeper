import Vue from 'vue'
import Router from 'vue-router'
import Layout from '../components/Layout'

Vue.use(Router)

export default new Router({
  mode: 'hash',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      redirect: '/config'
    },
    {
      path: '/config',
      component: Layout,
      children: [{
        name: 'config',
        path: '',
        component: () => import('../views/Config')
      }]
    },
    /*{
      path: '/keeper',
      component: Layout,
      children: [{
        name: 'keeper',
        path: '',
        component: () => import('../views/Keeper')
      }]
    }*/
  ]
})
