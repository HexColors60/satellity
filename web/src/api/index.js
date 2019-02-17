import forge from 'node-forge';
import moment from 'moment';
import KJUR from 'jsrsasign';
import uuid from 'uuid/v4';
import Cookies from 'js-cookie';
import axios from 'axios';
import Noty from 'noty';
import Category from './category.js';
import Comment from './comment.js';
import Topic from './topic.js';
import User from './user.js';

axios.defaults.baseURL = 'https://api.godiscourse.com';
if (process.env.NODE_ENV === 'development') {
  axios.defaults.baseURL = 'http://localhost:4000';
}
axios.defaults.headers.common['Content-Type'] = 'application/json';

axios.interceptors.request.use(function(config) {
  let method = config.method, url = config.url, data = config.data;
  config.headers.common['Authorization'] = `Bearer ${token(method, url, data)}`;
  return config
}, function(error) {
  return Promise.reject(error);
});

axios.interceptors.response.use(function(response) {
  if (!!response.status && response.status === 200) {
    let data = response.data;
    return data;
  }
  return response
}, function(error) {
  return Promise.reject(error);
});

function token(method, uri, body) {
  let priv = window.localStorage.getItem('token');
  let pwd = Cookies.get('sid');
  if (!priv || !pwd) {
    return "";
  }
  Cookies.set('sid', pwd, { expires: 365 });

  let uid = window.localStorage.getItem('uid');
  let sid = window.localStorage.getItem('sid');
  return sign(uid, sid, priv, method, uri, body);
}

function sign(uid, sid, privateKey, method, uri, body) {
  if (typeof body !== 'string') { body = ""; }

  let expire = moment.utc().add(30, 'minutes').unix();
  let md = forge.md.sha256.create();
  md.update(method + uri + body);

  let oHeader = {alg: 'ES256', typ: 'JWT'};
  let oPayload = {
    uid: uid,
    sid: sid,
    exp: expire,
    jti: uuid(),
    sig: md.digest().toHex()
  };
  let sHeader = JSON.stringify(oHeader);
  let sPayload = JSON.stringify(oPayload);
  let pwd = Cookies.get('sid');
  try {
    let k = KJUR.KEYUTIL.getKey(privateKey, pwd);
  } catch (e) {
    return '';
  }
  return KJUR.jws.JWS.sign('ES256', sHeader, sPayload, privateKey, pwd);
}

class API {
  constructor() {
    this.axios = axios;
    this.category = new Category(this);
    this.comment = new Comment(this);
    this.topic = new Topic(this);
    this.user = new User(this);
  }
}

export default API;
