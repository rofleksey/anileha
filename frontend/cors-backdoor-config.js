module.exports = {
  target: 'http://192.168.1.88:7891',
  baseUrl: '',
  port: 1234,
  headers: {
    'Access-Control-Allow-Origin': 'http://192.168.1.88:9000',
    'Access-Control-Allow-Methods': 'GET,POST,HEAD,METHOD,DELETE,OPTIONS,PUT',
    'Access-Control-Allow-Headers':
      'Origin, X-Requested-With, Content-Type, Accept, Authorization, Set-Cookie, Cookie',
    'Access-Control-Expose-Headers': 'x-some-custom-header',
    'Access-Control-Allow-Credentials': 'true',
    'x-powered-by': undefined,
  },
  debug: {
    name: 'cors-backdoor',
    level: 'trace',
    prettyPrint: true,
  },
};
