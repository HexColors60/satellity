import style from './style.css';
import "simplemde/dist/simplemde.min.css";
import SimpleMDE from 'react-simplemde-editor';
import React, { Component } from 'react';
import { Link } from 'react-router-dom';
import API from '../api/index.js';

class TopicNew extends Component {
  constructor(props) {
    super(props);
    this.api = new API();
    this.state = {title: '', category_id: '', body: '', categories: []};
    this.handleChange = this.handleChange.bind(this);
    this.handleBodyChange = this.handleBodyChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
    const classes = document.body.classList.values();
    document.body.classList.remove(...classes);
    document.body.classList.add('topic', 'layout');
    // TODO handle authentication
    if (!this.api.user.loggedIn()) {
      props.history.push('/');
    }
  }

  componentDidMount() {
    this.api.category.index((resp) => {
      let category_id = '';
      if (resp.data.length > 0) {
        category_id = resp.data[0].category_id;
      }
      this.setState({categories: resp.data, category_id: category_id});
    });
  }

  handleChange(e) {
    const target = e.target;
    const name = target.name;
    this.setState({
      [name]: target.value
    });
  }

  handleBodyChange(value) {
    this.setState({body: value});
  }

  handleSubmit(e) {
    e.preventDefault();
    const history = this.props.history;
    const data = {title: this.state.title, body: this.state.body, category_id: this.state.category_id};
    this.api.topic.create(data, (resp) => {
      history.push('/');
    });
  }

  render() {
    return (
      <View onSubmit={this.handleSubmit} onChange={this.handleChange} onBodyChange={this.handleBodyChange} state={this.state} />
    )
  }
}

// TODO jsx editor format
const View = ({onSubmit, onChange, onBodyChange, state}) => {
  const categories = state.categories.map((c) => {
    return (
      <option value={c.category_id} key={c.category_id}>{c.name}</option>
    )
  });

  return (
    <div className='container'>
      <main className='section main'>
        <div className={style.form}>
          <h2>Create a new topic</h2>
          <form onSubmit={(e) => onSubmit(e)}>
            <div>
              <label name='title'>Title *</label>
              <input type='text' name='title' pattern='.{3,}' required value={state.title} autoComplete='off' onChange={(e) => onChange(e)} />
            </div>
            <div>
              <label name='category'>Category</label>
              <div className='select'>
                <select name='category_id' value={state.category_id} onChange={(e) => onChange(e)}>
                  {categories}
                </select>
              </div>
            </div>
            <div>
              <SimpleMDE
                className={""}
                value={state.body}
                onChange={onBodyChange}
              />
            </div>
            <div className='action'>
              <input type='submit' value='SUBMIT' />
            </div>
          </form>
        </div>
      </main>
      <aside className='section aside'>
      </aside>
    </div>
  )
};

export default TopicNew;
