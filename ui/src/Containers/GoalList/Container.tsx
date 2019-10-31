import React from 'react';
import goalServices from './Services';

import EmptyBox from '../../Components/Variants/EmptyBox';
import LinkButton from '../../Components/Variants/LinkButton';

const NoGoals = () => (
  <EmptyBox>
    <p>You haven’t added any goals yet.</p>
    <LinkButton href='/create-goal' to='/create-goal'>
      Add a Goal
    </LinkButton>
  </EmptyBox>
);

interface GoalListProps {}

const GoalList = (props: GoalListProps) => {
  const [goalsList, setGoalsList] = React.useState([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  React.useEffect(async function {
    try {
      const { data } = await goalServices.getGoals();
      setGoalsList(data);
      setError(false);
    } catch(ex) {
      setError(true);
      console.error(ex);
    } finally {
      setLoading(false);
    }
  }, []);
  return (
    <EmptyBox>
      {loading === false && goalsList.length === 0 && <NoGoals />}
      {loading === false && error && ('invalid')}
      {loading === true && ('loading')}
      {loading === false && goalsList.length > 0}
    </EmptyBox>
  );
};

export default GoalList;
