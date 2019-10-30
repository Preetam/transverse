import React from 'react';
import { ThemeProvider } from 'styled-components';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link,
  Redirect
} from 'react-router-dom';

import { ChronicaProFontStyles, dark, light } from './Styles';

import CreateGoal from './Containers/CreateGoal';
import GoalList from './Containers/GoalList';
import GoalDetail from './Containers/GoalDetail';
import GoalPreferences from './Containers/GoalPreferences';
import Profile from './Containers/Profile';

export default function App() {
  let theme = light;
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)')) {
    theme = dark;
  }
  return (
    <ThemeProvider theme={theme}>
      <Router>
        <ChronicaProFontStyles />
        <Switch>
          <Route path='/goals'>
            <GoalList />
          </Route>
          <Route path='/create-goal'>
            <CreateGoal />
          </Route>
          <Route path='/goals/:id'>
            <GoalDetail />
          </Route>
          <Route path='/goals/:id/settings'>
            <GoalPreferences />
          </Route>
          <Route path='/profile'>
            <Profile />
          </Route>
          <Route path='*'>
            <Redirect to='/goals' />
          </Route>
        </Switch>
      </Router>
    </ThemeProvider>
  );
}
