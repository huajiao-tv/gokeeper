import vue from 'vue'
import axios from 'axios'

vue.prototype.$axios = axios
axios.defaults.baseURL = '/api'
axios.defaults.timeout = 5000
