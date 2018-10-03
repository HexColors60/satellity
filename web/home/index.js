import React, { Component } from 'react';
import { Link } from "react-router-dom";

class Home extends Component {
  constructor(props) {
    super(props);
    const classes = document.body.classList.values();
    document.body.classList.remove(...classes);
    document.body.classList.add('home', 'layout');
  }

  render() {
    return (
      <HomeView />
    );
  }
}

const HomeView = (match) => (
  <div>
    <div>Home</div>
    <Link to='/about'>About</Link>
  </div>
);

export default Home;
