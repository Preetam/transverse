import axios from 'axios';

const restClient = axios.create({
  baseURL: window.location.origin,
  timeout: 1000,
  headers: { 'X-Requested-With': 'XMLHttpRequest' }
});

export default restClient;
