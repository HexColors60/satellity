import './index.scss';
import Typed from 'typed.js';
import React, { Component } from 'react';
import { Link } from 'react-router-dom';
import API from '../api/index.js';

class Home extends Component {
  constructor(props) {
    super(props);
    const classes = document.body.classList.values();
    document.body.classList.remove(...classes);
    document.body.classList.add('home', 'layout');
  }

  componentDidMount() {
    const options = {
      strings: ['Welcome, this is Suntin', 'A discourse-like forum in Go', 'Welcome, Suntin!'],
      typeSpeed: 100,
      backSpeed: 50
    };
    this.typed = new Typed('.welcome', options);
  }

  componentWillUnmount() {
  }

  render() {
    return (
      <HomeView />
    );
  }
}

const HomeView = (match) => (
  <div className='container'>
    <span className='welcome'></span>
  </div>
);

export default Home;
