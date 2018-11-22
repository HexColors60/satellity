import './sign_in.scss';
import React, { Component } from 'react';
import { Link } from 'react-router-dom';

class SignIn extends Component {
  constructor(props) {
    super(props);
    const classes = document.body.classList.values();
    document.body.classList.remove(...classes);
    document.body.classList.add('sign_in', 'layout');
  }

  render() {
    let githubClientId = '71905afbd6e4541ad62b';
    if (process.env.NODE_ENV === 'development') {
      githubClientId = '03e10a9b62b4533e65b5';
    }
    return (
      <SignInView clientId={githubClientId} />
    );
  }
}

const SignInView = (props, match) => (
  <div>
    <h1 className='brand'><Link to='/'>Suntin</Link></h1>
    <div>
      <a href={`https://github.com/login/oauth/authorize?scope=user:email&client_id=${props.clientId}`} className='btn primary'>Sign in with GitHub</a>
    </div>
  </div>
);

export default SignIn;
