import style from './assets/css/not_fount.scss';
import React from 'react';
import { Link } from 'react-router-dom';

const NoMatch = ({ location }) => {
  const classes = document.body.classList.values();
  document.body.classList.remove(...classes);
  document.body.classList.add('not-found', 'layout');

  return (
    <div className={style.container}>
      <h3 className={style.body}>
         LOL! NO MATCH FOR <span>{location.pathname}</span>
        <div className={style.action}>
          <Link to='/'>Back to homepage</Link>
        </div>
      </h3>
    </div>
  )
};

export default NoMatch;
