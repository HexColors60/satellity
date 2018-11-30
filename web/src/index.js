import '../node_modules/normalize.css/normalize.css';
import '../node_modules/@fortawesome/fontawesome-free/css/all.css';
import './assets/css/h5bp.css';
import './index.scss';
import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter as Router, Route, Link, Switch } from 'react-router-dom';
import MainRoute from './layouts/main.js';
import AdminRoute from './layouts/admin.js';
import NoMatch from './notfound.js';
import Oauth from './users/oauth.js';
import SignIn from './users/sign_in.js';
import TopicNew from './topics/new.js';
import About from './about.js';
import Home from './home/index.js';

import { library } from '@fortawesome/fontawesome-svg-core';
import { faEdit, faTrashAlt } from '@fortawesome/free-regular-svg-icons';
library.add(faEdit, faTrashAlt);

ReactDOM.render((
  <Router>
    <div>
      <Switch>
        <MainRoute exact path='/' component={Home} />
        <MainRoute path='/about' component={About} />
        <MainRoute exact path='/topics/new' component={TopicNew} />
        <MainRoute path='/topics/:id/edit' component={About} />
        <MainRoute path='/topics/:id' component={About} />
        <Route path='/sign_in' component={SignIn} />
        <Route path='/oauth/callback' component={Oauth} />
        <Route path='/admin' component={AdminRoute} />
        <Route component={NoMatch} />
      </Switch>
    </div>
  </Router>
), document.querySelector('#layout-container'));
