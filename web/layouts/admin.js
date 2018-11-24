import '../admin/admin.scss';
import React, { Component } from 'react';
import { Route, Link } from 'react-router-dom';
import constants from '../components/constants.js';
import Index from '../admin/index.js';
import Categories from '../admin/categories/index.js';
import CategoriesNew from '../admin/categories/new.js';
import About from '../about.js';
import API from '../api/index.js';

class AdminRoute extends Component {
  constructor(props) {
    super(props);
    if (!new API().user.isAdmin()) {
      props.history.push('/');
      return
    }
    this.state = {site: constants.site}
    const classes = document.body.classList.values();
    document.body.classList.remove(...classes);
    document.body.classList.add('admin', 'layout');
  }

  render() {
    const match = this.props.match;
    return (
      <div>
        <header className='header navi'>
          <Link to='/' className='brand'> &larr; Back to {this.state.site}</Link>
          <Link to='/admin'>Dashboard</Link>
          <Link to='/admin/categories'>Categories</Link>
        </header>
        <div className='container'>
          <div className='wrapper'>
            <Route exact path={`${match.url}`} component={Index} />
            <Route exact path={`${match.url}/categories`} component={Categories} />
            <Route path={`${match.url}/categories/new`} component={CategoriesNew} />
            <Route path={`${match.url}/about`} component={About} />
          </div>
        </div>
      </div>
    )
  }
}

export default AdminRoute;
