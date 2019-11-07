import axios from 'axios';

const restClient = axios.create({
  baseURL: 'http://0.0.0.0:4001', //window.location.origin,
  timeout: 1000,
  headers: { 'X-Requested-With': 'XMLHttpRequest' }
});

export default restClient;
